package psql

import (
	"context"

	"time"

	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

type Psql struct {
	url        string
	dbpool     *pgxpool.Pool
	incomingCh chan interface{}
	cancel     func()
}

func InitPsql(url string, waitingTime int) (*Psql, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Psql{
		cancel: cancel,
	}

	// запуск процесса переподключения
	go p.processConn(ctx, waitingTime)
	go p.listenAndServe(ctx)

	return p, ctx
}

// процесс проверяет подключение к postgres в заданный промежуток
// если подключения нет то производится попытка подключения
func (p *Psql) processConn(ctx context.Context, waitingTime int) {
	defer log.Warning("processConn is closed")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := p.checkPool(); err != nil {
				err = p.connPool(context.TODO())
				if err != nil {
					log.Error(err)
				}
			}
			time.Sleep(time.Duration(waitingTime) * time.Second)
		}
	}
}

// метод для создания пула коннектов
func (p *Psql) connPool(ctx context.Context) error {
	dbpool, err := pgxpool.New(ctx, p.url)
	if err != nil {
		return err
	}
	p.dbpool = dbpool
	log.Info("a pool of PostgreSQL connections has been created")
	return err
}

func (p *Psql) checkPool() error {
	err := p.dbpool.Ping(context.TODO())
	return err
}

// процесс
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

}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (p *Psql) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	p.cancel()
}
