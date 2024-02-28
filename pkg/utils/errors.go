package utils

// ошибка приведения типов
type typeConversionError struct {
}

func (e typeConversionError) Error() string {
	return "type conversion error"
}
