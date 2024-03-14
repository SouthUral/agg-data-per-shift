package storage

import (
	"context"
	"fmt"
	"time"

	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

// структура объекта работы с БД
type PgConn struct {
	url            string
	timeOutQueryes time.Duration
	dbpool         *pgxpool.Pool
	cancel         func()
}

func getUrl(data map[string]string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%s&pool_min_conns=%s",
		data["user"],
		data["password"],
		data["host"],
		data["port"],
		data["db_name"],
		data["pool_max_conns"],
		data["pool_min_conns"],
	)
}

func InitPgConn(pgDataVars map[string]string, timeOutQueryes, checkTimeWait int) (*PgConn, context.Context) {

	ctx, cancel := context.WithCancel(context.Background())

	p := &PgConn{
		url:            getUrl(pgDataVars),
		cancel:         cancel,
		timeOutQueryes: time.Duration(timeOutQueryes),
	}

	// запуск процесса мониторинга и подключения к БД
	go p.processConn(ctx, checkTimeWait)

	return p, ctx
}

// процесс проверяет подключение к postgres в заданный промежуток
// если подключения нет то производится попытка подключения
func (p *PgConn) processConn(ctx context.Context, waitingTime int) {
	var maxAcquiredConns int32
	var minAcquiredConns int32

	defer log.Warning("processConn is closed")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := p.checkPool(); err != nil {
				log.Error(err)
				err = p.connPool(context.TODO())
				if err != nil {
					log.Error(err)
				}
			}

			status := p.dbpool.Stat()
			acCons := status.AcquiredConns()
			if maxAcquiredConns < acCons {
				maxAcquiredConns = acCons
			}

			if minAcquiredConns > acCons {
				minAcquiredConns = acCons
			}

			log.Debugf("PgConnStatus! maxAcquiredConns: %d, minAcquiredConns: %d", maxAcquiredConns, minAcquiredConns)
			time.Sleep(time.Duration(waitingTime) * time.Millisecond)
		}
	}
}

// метод для создания пула коннектов
func (p *PgConn) connPool(ctx context.Context) error {
	dbpool, err := pgxpool.New(ctx, p.url)
	if err != nil {
		return err
	}
	p.dbpool = dbpool
	log.Info("a pool of PostgreSQL connections has been created")
	return err
}

func (p *PgConn) checkPool() error {
	if p.dbpool == nil {
		return fmt.Errorf("коннект еще не создан")
	}
	err := p.dbpool.Ping(context.TODO())
	return err
}

func (p *PgConn) closePoolConn() {
	err := p.checkPool()
	if err != nil {
		log.Warning("the connection pool has already been closed")
		return
	}
	p.dbpool.Close()
	log.Info("connection pool is closed")
}

// метод для запросов в БД (любой запрос, который вернет данные)
func (p *PgConn) QueryDB(query string, args ...any) (pgx.Rows, error) {
	ctx, _ := context.WithTimeout(context.Background(), p.timeOutQueryes*time.Second)

	rows, err := p.dbpool.Query(ctx, query, args...)
	if err != nil {
		err = utils.Wrapper(queryDBError{}, err)
		log.Error(err)
	}
	return rows, err
}

// метод производит запрос, который вернет одну строку
func (p *PgConn) QueryRowDB(query string, args ...any) pgx.Row {
	ctx, _ := context.WithTimeout(context.Background(), p.timeOutQueryes*time.Second)

	row := p.dbpool.QueryRow(ctx, query, args...)
	return row
}

// метод производит запрос, который не должен ничего возвращать
func (p *PgConn) ExecQuery(query string, args ...any) error {
	ctx, _ := context.WithTimeout(context.Background(), p.timeOutQueryes*time.Second)
	_, err := p.dbpool.Exec(ctx, query, args...)
	return err
}

// метод прекращает работу модуля PgConn (завершает все активные горутины, разрывает коннект с БД)
func (p *PgConn) Shutdown(err error) {
	status := p.dbpool.Stat()
	log.Infof("PgConn; all cons: %d ; AcquiredConns: %d; IdleConns: %d", status.TotalConns(), status.AcquiredConns(), status.IdleConns())
	log.Errorf("PgConn is terminated for a reason: %s", err)
	p.cancel()
	p.closePoolConn()
}
