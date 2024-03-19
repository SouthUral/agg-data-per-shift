package core

import (
	"context"
	"errors"
	"time"

	amqp "agg-data-per-shift/internal/amqp/amqp_client"
	aggMileage "agg-data-per-shift/internal/services/aggMileageHours"
	storage "agg-data-per-shift/internal/storage"
	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

const (
	getStreamOffset = "GetOffset"
	pushEvent       = "InputMSG"
)

var (
	errRabbitShutdownError       = errors.New("the rabbit module has stopped working")
	errUncknowTypeRabbitMesError = errors.New("unknown message type from the rabbit module")
	errStorageShutdownError      = errors.New("the storage module has stopped working")
	errRouterShutdownError       = errors.New("the router module has stopped working")
	errConverRabbitMesError      = errors.New("error converting a message from rabbit")
	errPgConnShutdownError       = errors.New("pgConn has stopped working")
	errTimeMeterFinishError      = errors.New("timeMeter has stopped working")
)

func InitCore() {
	var storageCtx, pgCtx, rabbitCtx, routerCtx, timeMeterCtx context.Context

	envs, err := getEnvs()
	if err != nil {
		log.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	core := core{
		cancel: cancel,
	}

	core.timeMeter, timeMeterCtx = utils.InitProcessingTimeMeter()
	// инициализация подключения к базам
	core.rabbit = amqp.InitRabbit(envs.rbEnvs, 30)
	core.pgConn, pgCtx = storage.InitPgConn(envs.pgEnvs, 1000, 1000, 20)

	core.storage, storageCtx = storage.InitStorageMessageHandler(core.pgConn)
	core.router, routerCtx = aggMileage.InitEventRouter(core.storage.GetStorageCh(), 10, core.timeMeter)

	// начало прослушивания очереди
	rabbitCtx = core.rabbit.StartRb()

	go core.controlProcess(ctx, storageCtx, rabbitCtx, routerCtx, pgCtx, timeMeterCtx)
	go core.routingEvents(ctx)

}

type core struct {
	timeMeter    *utils.ProcessingTimeMeter
	rabbit       *amqp.Rabbit
	storage      *storage.StorageMessageHandler
	pgConn       *storage.PgConn
	router       *aggMileage.EventRouter
	streamOffset string
	cancel       func()
}

func (c *core) shudown(err error) {
	c.storage.Shutdown(err)
	c.rabbit.RabbitShutdown(err)
	c.router.Shudown(err)
	c.pgConn.Shutdown(err)
	c.timeMeter.Shudown()
	c.cancel()
	log.Errorf("the program stopped working due to: %s", err.Error())
}

func (c *core) routingEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.rabbit.GetChan():
			message, ok := msg.(msgEvent)
			if !ok {
				c.shudown(errConverRabbitMesError)
				return
			}
			c.routingRabbitMes(message)

		}
	}
}

func (c *core) routingRabbitMes(mes msgEvent) {
	switch mes.GetTypeMsg() {
	case getStreamOffset:
		switch c.streamOffset {
		case "first":
			mes.GetReverceCh() <- answerEvent{
				offset: 100000000,
			}
		default:
			c.storage.GetStorageCh() <- transportStruct{
				sender:         "amqp",
				mesage:         mes.GetTypeMsg(),
				reverseChannel: mes.GetReverceCh(),
			}
		}
	case pushEvent:
		mes.GetReverceCh() <- answerEvent{}
		err := c.router.EventReception(mes)
		if err != nil {
			c.shudown(err)
			return
		}
	default:
		c.shudown(errUncknowTypeRabbitMesError)
	}

}

func (c *core) controlProcess(ctx, ctxStorage, ctxRabbit, ctxRouter, ctxPgConn, ctxTimeMeter context.Context) {
	defer log.Warning("core control process has finished")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ctxStorage.Done():
			c.shudown(errStorageShutdownError)
		case <-ctxRabbit.Done():
			c.shudown(errRabbitShutdownError)
		case <-ctxRouter.Done():
			c.shudown(errRouterShutdownError)
		case <-ctxPgConn.Done():
			c.shudown(errPgConnShutdownError)
		case <-ctxTimeMeter.Done():
			c.shudown(errTimeMeterFinishError)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
