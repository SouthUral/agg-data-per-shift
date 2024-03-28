package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

var (
	errAttemptConnError = errors.New("the number of attempts to connect to the database has ended")
)

// структура объекта работы с БД
type PgConn struct {
	url                       string
	timeOutQueryes            time.Duration
	dbpool                    *pgxpool.Pool
	intervalBetweenRequest    int // начальный интервал между повтором запросов
	maxIntervalBetweenRequest int // максимальный интервал между повтором запросов
	cancel                    func()
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

// Функция инициализирует PgConn и запускает процесс подключения к БД и мониторинга состояния подключения.
//   - pgDataVars: параметры загруженные из переменных окружения;
//   - timeOutQueryes: время ожидания ответа на запросы в БД (в миллисекундах);
//   - checkTimeWait: время ожидания между проверками подключения к БД (в миллисекундах);
//   - numAttemptConn: количество попыток реконнекта в случае дисконнекта БД;
//   - intervalBetweenRequest:  начальный интервал между повтором запросов;
//   - maxIntervalBetweenRequest:  максимальный интервал между повтором запросов
func InitPgConn(pgDataVars map[string]string, timeOutQueryes, checkTimeWait, numAttemptConn, intervalBetweenRequest, maxIntervalBetweenRequest int) (*PgConn, context.Context) {

	ctx, cancel := context.WithCancel(context.Background())

	p := &PgConn{
		url:                       getUrl(pgDataVars),
		cancel:                    cancel,
		timeOutQueryes:            time.Duration(timeOutQueryes),
		intervalBetweenRequest:    intervalBetweenRequest,
		maxIntervalBetweenRequest: maxIntervalBetweenRequest,
	}

	// запуск процесса мониторинга и подключения к БД
	go p.processConn(ctx, checkTimeWait, numAttemptConn)

	return p, ctx
}

// процесс проверяет подключение к postgres в заданный промежуток
// если подключения нет то производится попытка подключения
func (p *PgConn) processConn(ctx context.Context, waitingTime, numAttemptConn int) {
	var counterAttemptConn = numAttemptConn

	defer log.Warning("processConn is closed")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := p.checkPool(); err != nil {
				log.Error(err)
				if err = p.connPool(); err != nil {
					log.Error(err)
				}
				if counterAttemptConn == 0 {
					err = utils.Wrapper(errAttemptConnError, err)
					p.Shutdown(err)
					return
				}
				counterAttemptConn--
				log.Infof("there are still attempts to connect: %d", counterAttemptConn)
			} else {
				counterAttemptConn = numAttemptConn
				p.defineNumsConn()
			}

			time.Sleep(time.Duration(waitingTime) * time.Millisecond)
		}
	}
}

// метод для создания пула коннектов
func (p *PgConn) connPool() error {
	ctx, _ := context.WithTimeout(context.Background(), p.timeOutQueryes*time.Second)
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
		return fmt.Errorf("the pool of connections has not been created yet")
	}
	ctx, _ := context.WithTimeout(context.Background(), p.timeOutQueryes*time.Second)
	err := p.dbpool.Ping(ctx)
	if err != nil {
		log.Errorf("checkConn err: %s", err)
	}
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
func (p *PgConn) QueryDB(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	var (
		rows                    pgx.Rows
		err                     error
		intervalBetweenRequests = p.intervalBetweenRequest
	)
	for {
		select {
		case <-ctx.Done():
			return rows, err
		default:
			if p.checkPool() == nil {
				ctxQuery, _ := context.WithTimeout(ctx, p.timeOutQueryes*time.Second)
				rows, err = p.dbpool.Query(ctxQuery, query, args...)

				if err == nil {
					return rows, err
				}

				if isCriticalPgxError(err) {
					p.Shutdown(err)
					return rows, err
				}
			}

			intervalBetweenRequests = cycleRetarder(intervalBetweenRequests, p.maxIntervalBetweenRequest)
			log.Debug("Повторная попытка запроса")
		}
	}
}

// метод производит запрос, который вернет одну строку
func (p *PgConn) QueryRowDB(ctx context.Context, query string, args ...any) pgx.Row {
	ctxQuery, _ := context.WithTimeout(ctx, p.timeOutQueryes*time.Second)
	row := p.dbpool.QueryRow(ctxQuery, query, args...)
	return row
}

func (p *PgConn) QueryRowWithResponseInt(ctx context.Context, query string, responce *int, args ...any) error {
	var (
		err                     error
		intervalBetweenRequests = p.intervalBetweenRequest
	)

	for {
		select {
		case <-ctx.Done():
			return err
		default:
			err = p.QueryRowDB(ctx, query, args...).Scan(responce)
			if err == nil {
				return err
			}
			if isCriticalPgxError(err) {
				p.Shutdown(err)
				return err
			}
			intervalBetweenRequests = cycleRetarder(intervalBetweenRequests, p.maxIntervalBetweenRequest)
			log.Debug("Повторная попытка запроса")
		}
	}
}

// метод производит запрос, который не должен ничего возвращать
func (p *PgConn) ExecQuery(ctx context.Context, query string, args ...any) error {
	var (
		err                     error
		intervalBetweenRequests = p.intervalBetweenRequest
	)
	for {
		select {
		case <-ctx.Done():
			return err
		default:
			ctxQuery, _ := context.WithTimeout(ctx, p.timeOutQueryes*time.Second)
			_, err = p.dbpool.Exec(ctxQuery, query, args...)
			if err == nil {
				return err
			}
			if isCriticalPgxError(err) {
				p.Shutdown(err)
				return err
			}
			intervalBetweenRequests = cycleRetarder(intervalBetweenRequests, p.maxIntervalBetweenRequest)
			log.Debug("Повторная попытка запроса")
		}

	}
}

// выводит количество используемых и неиспользуемых коннектов
func (p *PgConn) defineNumsConn() {
	var maxAcquiredConns int32
	var minAcquiredConns int32

	status := p.dbpool.Stat()
	acCons := status.AcquiredConns()
	if maxAcquiredConns < acCons {
		maxAcquiredConns = acCons
	}

	if minAcquiredConns > acCons {
		minAcquiredConns = acCons
	}

	log.Debugf("PgConnStatus! maxAcquiredConns: %d, minAcquiredConns: %d", maxAcquiredConns, minAcquiredConns)
}

// метод прекращает работу модуля PgConn (завершает все активные горутины, разрывает коннект с БД)
func (p *PgConn) Shutdown(err error) {
	log.Errorf("PgConn is terminated for a reason: %s", err)
	p.cancel()
	p.closePoolConn()
}
