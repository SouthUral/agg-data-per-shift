package aggmileagehours

// ошибка десериализации json
type unmarshalingJsonError struct {
}

func (e unmarshalingJsonError) Error() string {
	return "unmarshaling json error"
}

// ошибка приведения типов
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
