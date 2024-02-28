package psql

import (
	"context"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type Psql struct {
	dbConn     *pgConn
	incomingCh chan interface{}
	cancel     func()
}

func InitPsql(url string, waitingTime int) (*Psql, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Psql{
		cancel: cancel,
		dbConn: initPgConn(url),
	}

	go p.listenAndServe(ctx)

	return p, ctx
}

// процесс получения и обработки сообщений от других модулей
func (p *Psql) listenAndServe(ctx context.Context) {
	defer log.Warning("listenAndServe is closed")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-p.incomingCh:
			message, err := utils.TypeConversion[trunsportMes](msg)
			if err != nil {
				err = utils.Wrapper(listenAndServeError{}, err)
				p.Shutdown(err)
			}
			p.handleRequests(message)
		}
	}

}

// обработчик запросов
func (p *Psql) handleRequests(message trunsportMes) {
	switch message.GetSender() {
	case aggMileageHours:
		// обработка сообщений от модуля aggMileageHours
	default:
		log.Errorf("unknown sender: %s", message.GetSender())
	}
}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (p *Psql) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	p.cancel()
	p.dbConn.shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
}
