package aggmileagehours

import (
	"errors"
	"fmt"
	"time"
)

var (
	errActiveEventRouter = errors.New("event router is not active")
	// ошибка обработчика (флаг не активен)
	errHadlerEventError          = errors.New("handler error")
	errHandlerNotRespondingError = errors.New("the handler is not responding")
	// ошибка восстановления состояния AggDataPerObject
	errRestoringStateError = errors.New("restoring state AggDataPerObject error")
	// ошибка добавления новой смены и сессии
	errAddNewShiftAndSessionError = errors.New("error adding a new shift and a new session")
	// ошибка добвления новой записи сессии и обновления записи смены
	errUpdateShiftAndAddNewSessionError = errors.New("error updating the shift record and adding a new session record")
	// ошибка обновления смены и сессии
	errUpdateShiftAndSessionError = errors.New("shift and session update error")
)

// ошибка десериализации json
type unmarshalingJsonError struct {
}

func (e unmarshalingJsonError) Error() string {
	return "unmarshaling json error"
}

// ошибка приведения типа
type typeConversionError struct {
}

func (e typeConversionError) Error() string {
	return "type conversion error"
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

// ошибка определения смены (нет смены)
type defineShiftError struct {
	eventTime time.Time
}

func (e defineShiftError) Error() string {
	return fmt.Sprintf("define shift error, there is no shift : %s", e.eventTime)
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

// ошибка активности обработчика
type aggObjIsNotActiveError struct {
	numObj int
}

func (e aggObjIsNotActiveError) Error() string {
	return fmt.Sprintf("aggObj with obj id: %d is not active", e.numObj)
}
