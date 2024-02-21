package aggmileagehours

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

	eventData, err = data.getDecryptedData()
	if err != nil {
		err = fmt.Errorf("%w: %w", timeParseError{}, err)
	}

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

func timeConversion(date string) (time.Time, error) {
	dateSplits := strings.Split(date, ".")
	dateSplits[1] = dateSplits[1][:3]
	timeToFormat := strings.Join(dateSplits, ".")
	resTime, err := time.Parse(timeLayOut, timeToFormat)
	return resTime, err
}
