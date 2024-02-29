package storage

import (
	"context"
	"time"

	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

// структура объекта работы с БД
type pgConn struct {
	url            string
	timeOutQueryes time.Duration
	dbpool         *pgxpool.Pool
	cancel         func()
}

func initPgConn(url string, timeOutQueryes int) *pgConn {
	ctx, cancel := context.WithCancel(context.Background())

	p := &pgConn{
		url:            url,
		cancel:         cancel,
		timeOutQueryes: time.Duration(timeOutQueryes),
	}

	// запуск процесса мониторинга и подключения к БД
	go p.processConn(ctx, 5)

	return p
}

// процесс проверяет подключение к postgres в заданный промежуток
// если подключения нет то производится попытка подключения
func (p *pgConn) processConn(ctx context.Context, waitingTime int) {
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
func (p *pgConn) connPool(ctx context.Context) error {
	dbpool, err := pgxpool.New(ctx, p.url)
	if err != nil {
		return err
	}
	p.dbpool = dbpool
	log.Info("a pool of PostgreSQL connections has been created")
	return err
}

func (p *pgConn) checkPool() error {
	err := p.dbpool.Ping(context.TODO())
	return err
}

func (p *pgConn) closePoolConn() {
	err := p.checkPool()
	if err != nil {
		log.Warning("the connection pool has already been closed")
		return
	}
	p.dbpool.Close()
	log.Info("connection pool is closed")
}

// метод для запросов в БД (любой запрос, который вернет данные)
func (p *pgConn) queryDB(query string, args ...any) (pgx.Rows, error) {
	ctx, _ := context.WithTimeout(context.Background(), p.timeOutQueryes*time.Second)

	rows, err := p.dbpool.Query(ctx, query, args)
	if err != nil {
		err = utils.Wrapper(queryDBError{}, err)
		log.Error(err)
	}
	return rows, err
}

// метод прекращает работу модуля pgConn (завершает все активные горутины, разрывает коннект с БД)
func (p *pgConn) shutdown(err error) {
	log.Errorf("pgConn is terminated for a reason: %s", err)
	p.cancel()
	p.closePoolConn()
}
