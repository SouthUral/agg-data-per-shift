package storage

import (
	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5"
)

// функция преобразует результат запроса в виде pgx.Rows в структуру типа T.
// функция обрабатывает только одну строку из запроса.
func convertingQueryRowResultIntoStructure[T any](rows pgx.Rows) (T, error) {
	var rowObj T

	rowObj, err := pgx.CollectOneRow[T](rows, pgx.RowToStructByName[T])
	if err != nil {
		err = utils.Wrapper(convertRowToStructError{}, err)
	}

	return rowObj, err
}
