package storage

import (
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

// функция преобразует результат запроса в виде pgx.Rows в структуру типа T.
// функция обрабатывает только одну строку из запроса.
func convertingQueryRowResultIntoStructure[T any](rows pgx.Rows) (T, error) {
	var rowObj T

	rowObj, err := pgx.CollectOneRow[T](rows, pgx.RowToStructByName[T])
	if err != nil {
		log.Error(err)
		return rowObj, err
	}

	return rowObj, err
}
