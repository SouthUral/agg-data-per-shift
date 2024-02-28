package psql

import (
	"context"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type StorageMessageHandler struct {
	dbConn     *pgConn
	incomingCh chan interface{}
	cancel     func()
}

func InitPsql(url string, waitingTime int) (*StorageMessageHandler, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &StorageMessageHandler{
		cancel: cancel,
		dbConn: initPgConn(url),
	}

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
			s.handleRequests(message)
		}
	}

}

// обработчик запросов
func (s *StorageMessageHandler) handleRequests(message trunsportMes) {
	switch message.GetSender() {
	case aggMileageHours:
		// обработка сообщений от модуля aggMileageHours
	default:
		log.Errorf("unknown sender: %s", message.GetSender())
	}
}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (s *StorageMessageHandler) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	s.cancel()
	s.dbConn.shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
}
