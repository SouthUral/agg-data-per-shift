package storage

import (
	"context"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type StorageMessageHandler struct {
	dbConn     *PgConn
	incomingCh chan interface{}
	cancel     func()
	amqpHandler
	aggMileageAndHoursHandler
}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (s *StorageMessageHandler) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	s.cancel()
	s.dbConn.Shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
}

func (s *StorageMessageHandler) GetStorageCh() chan interface{} {
	return s.incomingCh
}

func InitStorageMessageHandler(pgDataVars map[string]string) (*StorageMessageHandler, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &StorageMessageHandler{
		cancel:     cancel,
		dbConn:     initPgConn(pgDataVars, 10),
		incomingCh: make(chan interface{}),
	}

	s.amqpHandler.dbConn = s.dbConn
	s.aggMileageAndHoursHandler.dbConn = s.dbConn

	go s.listenAndServe(ctx)

	return s, ctx
}

// процесс получения и обработки сообщений от других модулей
func (s *StorageMessageHandler) listenAndServe(ctx context.Context) {
	defer log.Warning("listenAndServe is closed")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.incomingCh:
			message, err := utils.TypeConversion[trunsportMes](msg)
			if err != nil {
				err = utils.Wrapper(listenAndServeError{}, err)
				s.Shutdown(err)
			}
			s.handleRequests(ctx, message)
		}
	}

}

// обработчик запросов
func (s *StorageMessageHandler) handleRequests(ctx context.Context, message trunsportMes) {
	switch message.GetSender() {
	case aggMileageHours:
		go s.handlerMesAggMileageHours(ctx, message)
	case amqp:
		go s.handlerMsgFromAmqp(message)
	default:
		log.Errorf("unknown sender: %s", message.GetSender())
	}
}
