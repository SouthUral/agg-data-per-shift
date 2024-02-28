package aggmileagehours

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	utils "agg-data-per-shift/pkg/utils"

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

	storageAnswer, err := utils.TypeConversion[incomingMessageFromStorage](answer)
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

// функция сравнивания двух дат
func comparingDates(dateFirst, dateSecond time.Time) bool {
	t1 := time.Date(dateFirst.Year(), dateFirst.Month(), dateFirst.Day(), 0, 0, 0, 0, time.Local)
	t2 := time.Date(dateSecond.Year(), dateSecond.Month(), dateSecond.Day(), 0, 0, 0, 0, time.Local)
	return t1.Equal(t2)
}
