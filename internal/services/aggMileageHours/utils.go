package aggmileagehours

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// функция преобразования []byte во внутреннюю структуру eventData
func decodingMessage(msg []byte) (*eventData, error) {
	var eventData *eventData

	data := &rawEventData{}
	start := time.Now()
	err := json.Unmarshal(msg, data)
	duration := time.Since(start)
	log.Debugf("время преобразования %d", duration.Microseconds())
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

// функция для преобразования ответа от модуля storage во интерфейсы модуля
func сonversionAnswerStorage(answer interface{}) (storageAnswerData, error) {
	var err error
	convertedStorageData := storageAnswerData{}

	storageAnswer, err := typeConversion[incomingMessageFromStorage](answer)
	if err != nil {
		err = fmt.Errorf("%w: %w", typeConversionAnswerStorageDataError{}, err)
		return convertedStorageData, err
	}

	err = convertedStorageData.shiftData.loadingInterfaceData(storageAnswer.GetDataShift())
	if err != nil {
		err = fmt.Errorf("%w: %w", typeConversionAnswerStorageDataError{}, err)
		return convertedStorageData, err
	}

	err = convertedStorageData.driverSessionData.loadingInterfaceData(storageAnswer.GetDataDriverSession())
	if err != nil {
		err = fmt.Errorf("%w: %w", typeConversionAnswerStorageDataError{}, err)
		return convertedStorageData, err
	}

	return convertedStorageData, err
}
