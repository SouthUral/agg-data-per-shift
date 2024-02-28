package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

// функция дженерик, для выполнения любого запроса, который должен вернуть одну строку,
// результат запроса обязательно должен быть передан в структуру T
func queryRow[T any](pgConn *pgConn, query string, args ...any) (T, error) {
	var rowObj T

	ctx, _ := context.WithTimeout(context.Background(), pgConn.timeOutQueryes*time.Second)

	rows, err := pgConn.dbpool.Query(ctx, query, args)
	if err != nil {
		log.Error(err)
		return rowObj, err
	}

	rowObj, err = pgx.CollectOneRow[T](rows, pgx.RowToStructByName[T])
	if err != nil {
		log.Error(err)
		return rowObj, err
	}

	return rowObj, err
}
