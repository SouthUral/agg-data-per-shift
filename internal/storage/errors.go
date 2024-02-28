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
