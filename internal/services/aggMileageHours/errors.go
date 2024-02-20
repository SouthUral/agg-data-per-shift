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

// остановка EventRouter
type stoppedEventRouterError struct {
}

func (e stoppedEventRouterError) Error() string {
	return "eventRouter has stopped working for a reason:"
}
