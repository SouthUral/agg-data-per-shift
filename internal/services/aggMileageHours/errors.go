package aggmileagehours

import "fmt"

// ошибка десериализации json
type unmarshalingJsonError struct {
}

func (e unmarshalingJsonError) Error() string {
	return "unmarshaling json error"
}

// ошибка приведения типа ответа от модуля storage
type typeConversionAnswerStorageDataError struct {
}

func (e typeConversionAnswerStorageDataError) Error() string {
	return "type conversion storage data error"
}

// остановка EventRouter
type stoppedEventRouterError struct {
}

func (e stoppedEventRouterError) Error() string {
	return "eventRouter has stopped working for a reason:"
}

// остановка EventRouter
type timeParseError struct {
}

func (e timeParseError) Error() string {
	return "time parse error"
}

// закончилось ожидание ответа от БД
type timeOutWaitAnswerDBError struct {
}

func (e timeOutWaitAnswerDBError) Error() string {
	return "the waiting time for a response from DB has ended"
}

// закрылся контекст AggDataPerObject
type contextAggPerObjectClosedError struct {
}

func (e contextAggPerObjectClosedError) Error() string {
	return "the AggDataPerObject context has been closed"
}

// ошибка преобразования интерфейса в указанную структуру
type interfaceLoadingToStructError struct {
	structName string
}

func (e interfaceLoadingToStructError) Error() string {
	return fmt.Sprintf("%s interface loading error", e.structName)
}

// ошибка восстановления состояния AggDataPerObject
type restoringStateError struct {
}

func (e restoringStateError) Error() string {
	return "restoring state AggDataPerObject error"
}

// ошибка определения смены (нет смены)
type defineShiftError struct {
}

func (e defineShiftError) Error() string {
	return "define shift error, there is no shift"
}

// ошибка при создании новой смены и сессии
type createNewObjectsError struct {
}

func (e createNewObjectsError) Error() string {
	return "createNewObjects error"
}

// ошибка при отправке и обработке сообщений от модуля storage
type processAndSendToStorageError struct {
}

func (e processAndSendToStorageError) Error() string {
	return "processAndSendToStorage error"
}

// количество попыток отправок запросов закончилось
type attemptRequestError struct {
}

func (e attemptRequestError) Error() string {
	return "the number of attempts to send requests has ended"
}
