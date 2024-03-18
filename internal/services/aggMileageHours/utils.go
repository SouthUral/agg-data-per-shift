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
	log.Warningf("timeConversion data: %s", date)
	dateSplits := strings.Split(date, ".")
	if len(dateSplits) == 1 {
		dateSplits = append(dateSplits, "000")
	}
	if len(dateSplits[1]) < 3 {
		dateSplits[1] = dateSplits[1] + "000000"
	}
	log.Warningf("%s", dateSplits[1])
	dateSplits[1] = dateSplits[1][:3]
	timeToFormat := strings.Join(dateSplits, ".")
	resTime, err := time.Parse(timeLayOut, timeToFormat)
	return resTime, err
}

// функция для преобразования ответа от модуля storage в интерфейс
func сonversionAnswerStorage[T any](responce interface{}) (T, error) {

	storageAnswer, err := utils.TypeConversion[T](responce)
	if err != nil {
		err = utils.Wrapper(typeConversionAnswerStorageDataError{}, err)
	}

	return storageAnswer, err
}

// приводит переданный интерфейс к указанному типу
func typeСonversion[T any](inputInterface interface{}) (T, error) {
	var err error
	resTypeData, ok := inputInterface.(T)
	if !ok {
		err = typeConversionError{}
	}
	return resTypeData, err
}

// функция сравнивания двух дат
func comparingDates(dateFirst, dateSecond time.Time) bool {
	t1 := time.Date(dateFirst.Year(), dateFirst.Month(), dateFirst.Day(), 0, 0, 0, 0, time.Local)
	t2 := time.Date(dateSecond.Year(), dateSecond.Month(), dateSecond.Day(), 0, 0, 0, 0, time.Local)
	return t1.Equal(t2)
}

// функция получает критические ошибки и возвращает первую найденную
func checkErrorsInMes(data incomingMessageFromStorage) error {
	var errCritical error

	if errCritical = data.GetCriticalErr(); errCritical != nil {
		return errCritical
	}
	if errCritical, _ := data.GetErrorsResponceShift(); errCritical != nil {
		return errCritical
	}
	errCritical, _ = data.GetErrorsResponceSession()
	return errCritical
}
