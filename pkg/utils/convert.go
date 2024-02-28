package utils

// функция преобразования типа
func TypeConversion[T any](data interface{}) (T, error) {
	var err error

	conversionData, ok := data.(T)
	if !ok {
		err = typeConversionError{}
	}

	return conversionData, err
}
