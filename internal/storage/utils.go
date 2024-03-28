package storage

import (
	"errors"
	"time"

	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// функция для конвертации rows в одну структуру
func converQuery[T any](rows pgx.Rows) (T, error) {
	var rowObj T

	rowObj, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		noRows := noRowsError{}
		if err.Error() == noRows.Error() {
			err = utils.Wrapper(convertRowToStructError{}, noRows)
		}
	}

	return rowObj, err
}

// функция для конвертации данных бд в слайс структур
func convertQueryRows[T any](rows pgx.Rows) ([]T, error) {
	var rowsObj []T
	var err error

	rowsObj, err = pgx.CollectRows[T](rows, pgx.RowToStructByName[T])
	if err != nil {
		err = utils.Wrapper(convertRowToStructError{}, err)

	}
	return rowsObj, err
}

// функция определяет критическая ли ошибка  pgx, если ошибка критическая то вернen true
func isCriticalPgxError(err error) bool {
	var pgErr *pgconn.PgError
	var pgConnError *pgconn.ConnectError
	if errors.As(err, &pgConnError) {
		return false
	}
	if errors.As(err, &pgErr) {
		return isCriticalCodePgError(pgErr.Code)
	}
	return true
}

// функция определяет, является ли критическим код ошибки
func isCriticalCodePgError(code string) bool {
	switch code[:2] {
	case "57":
		return false
	default:
		return true
	}
}

// функция замедитель, нужна для замедления повторяющихся запросов в цикле.
//   - timeSleep: время сна (в миллисекундах);
//   - maxTimeSleep: максимальное время сна  (в миллисекундах);
func cycleRetarder(timeSleep, maxTimeSleep int) int {
	time.Sleep(time.Duration(timeSleep) * time.Millisecond)
	if timeSleep < maxTimeSleep {
		return timeSleep * 2
	}
	return maxTimeSleep
}
