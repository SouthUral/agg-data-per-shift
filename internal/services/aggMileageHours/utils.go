package aggmileagehours

import (
	"encoding/json"
	"fmt"
)

// функция преобразования []byte во внутреннюю структуру eventData
func decodingMessage(msg []byte) (*eventData, error) {
	var eventData *eventData

	data := &rawEventData{}

	err := json.Unmarshal(msg, data)
	if err != nil {
		err = fmt.Errorf("%w: %w", unmarshalingJsonError{}, err)
		return eventData, err
	}

	eventData = data.getDecryptedData()

	return eventData, err
}

// функция преобразования типа
func typeConversion[T any](data interface{}) (T, error) {
	var err error

	conversionData, ok := data.(T)
	if !ok {
		err = typeConversionError{}
	}

	return conversionData, err
}
