package storage

import (
	utils "agg-data-per-shift/pkg/utils"
	"context"

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
func (a *amqpHandler) handlerMsgFromAmqp(ctx context.Context, msg trunsportMes) {
	message, err := utils.TypeConversion[string](msg.GetMesage())
	if err != nil {
		return
	}

	switch message {
	case getOffsetForAmqp:
		offset, err := a.getStreamLastOffset(ctx)
		if err != nil {
			err = utils.Wrapper(getStreamsOffsetError{}, err)
			log.Error(err)
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
func (a *amqpHandler) getStreamLastOffset(ctx context.Context) (int, error) {
	var offset int
	// row := a.dbConn.QueryRowDB(getStreamOffset)
	// err := row.Scan(&offset)
	err := a.dbConn.QueryRowWithResponseInt(ctx, getStreamOffset, &offset)
	return offset, err
}

// ответное сообщение
type answerEvent struct {
	offset int
}

func (a answerEvent) GetOffset() int {
	return a.offset
}
