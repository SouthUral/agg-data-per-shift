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

// проверка на наличие ошибки при произведении и обработке запросов
func handlingErrors(errors ...error) (string, error) {
	// критическая ошибка имеет наивысший приоритет, если выйдет хотя бы одна критическая ошибка, вторая ошибка уже обрабатываться не будет, программа сразу завершится
	// если хотя бы одна ошибка modulError то она будет передана в ответе
	// если обе ошибки regular то никакие дейсвия предприниматься не будут, ошибка не будет передана в модуль, программа не будет завершена
	var typeCriticalErr string
	var err error
	for _, err = range errors {
		if err == nil {
			continue
		}
		typeCriticalErr, err = defineTypeErrors(err)
		switch typeCriticalErr {
		case criticalError:
			return typeCriticalErr, err
		case modulError:
			return modulError, err
		default:
			continue
		}
	}
	return typeCriticalErr, err
}

// функция обработки ошибок, возвращает ошибку и ее определение, передать на обработку модулю который сделал запрос,
// обработать ошибку на месте, или завершить работу (если ошибка критическая)
func defineTypeErrors(err error) (string, error) {
	switch {
	case errors.Is(err, noRowsError{}):
		return modulError, err
	default:
		return criticalError, err
	}
}
