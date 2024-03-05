package storage

type listenAndServeError struct {
}

func (e listenAndServeError) Error() string {
	return "error in listenAndServe"
}

// ошибка в процессе обработки сообщения от модуля aggMileageHours
type handlerMesAggMileageHoursError struct {
}

func (e handlerMesAggMileageHoursError) Error() string {
	return "error in processing a message from the aggMileageHours module"
}

// ошибка запроса к БД
type queryDBError struct {
}

func (e queryDBError) Error() string {
	return "query DB error"
}

// ошибка конвертирования строки ответа на запрос в структуру
type convertRowToStructError struct {
}

func (e convertRowToStructError) Error() string {
	return "error convert row to struct"
}

type responceShiftSessionError struct {
}

func (e responceShiftSessionError) Error() string {
	return "responce shift or session error"
}

type noRowsError struct {
}

func (e noRowsError) Error() string {
	return "no rows in result set"
}
