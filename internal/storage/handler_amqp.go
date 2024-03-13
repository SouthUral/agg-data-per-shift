package storage

import (
	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type getStreamsOffsetError struct {
}

func (e getStreamsOffsetError) Error() string {
	return "error receiving stream offset from the database"
}

// структура для доступа к обработчику сообщений от amqp
type amqpHandler struct {
	dbConn *PgConn
}

// метод запускается как горутина
func (a *amqpHandler) handlerMsgFromAmqp(msg trunsportMes) {
	message, err := utils.TypeConversion[string](msg.GetMesage())
	if err != nil {
		// TODO: нужно отправить ошибку в распределитель сообщений
		return
	}
	switch message {
	case getOffsetForAmqp:
		offset, err := a.getStreamLastOffset()
		if err != nil {
			err = utils.Wrapper(getStreamsOffsetError{}, err)
			log.Error(err)
			// TODO: нужно отправить ошибку в распределитель сообщений
			return
		}
		log.Debugf("Получен последний оффсет из БД %d", offset)
		msg.GetChForResponse() <- answerEvent{
			offset: offset,
		}
	default:
		log.Warning("неизвестное сообщение от amqp")
		return
	}
}

// метод получения stream`s offset
func (a *amqpHandler) getStreamLastOffset() (int, error) {
	var offset int
	row := a.dbConn.QueryRowDB(getStreamOffset)
	err := row.Scan(&offset)
	return offset, err
}

// ответное сообщение
type answerEvent struct {
	offset int
}

func (a answerEvent) GetOffset() int {
	return a.offset
}
