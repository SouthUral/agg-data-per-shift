package storage

import (
	"errors"

	utils "agg-data-per-shift/pkg/utils"

	"github.com/jackc/pgx/v5"
)

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

// принимает одну или две любые ошибки, определяет критические и регулярные ошибки.
//   - первая ошибка - это критическая ошибка (с точки зрения storage);
//   - вторая ошибка - регулярная (могут быть предприняты иные действия).
func handlingErrors(errors ...error) (error, error) {
	var criticalErr, commonErr error
	var typeError string

	for _, err := range errors {
		if err != nil {
			typeError, err = defineTypeErrors(err)
			switch typeError {
			case criticalError:
				criticalErr = err
			case commonError:
				commonErr = err
			}
		}
	}
	return criticalErr, commonErr
}

// функция обработки ошибок, возвращает ошибку и ее определение, передать на обработку модулю который сделал запрос,
// обработать ошибку на месте, или завершить работу (если ошибка критическая)
func defineTypeErrors(err error) (string, error) {
	switch {
	case errors.Is(err, noRowsError{}):
		return commonError, err
	default:
		return criticalError, err
	}
}
