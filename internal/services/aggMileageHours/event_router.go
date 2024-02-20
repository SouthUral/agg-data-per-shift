package aggmileagehours

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// структура в которой происходит распределение событий по горутинам объектов техники
// ---
type EventRouter struct {
	incomingEventCh chan interface{}          // канал для получения событий
	storageCh       chan interface{}          // канал для связи с модулем psql
	aggObjs         map[int]*AggDataPerObject // map с объектами агрегации данных (по id техники)
	settingShift    *settingsDurationShifts   // настройки смены
	cancel          func()
}

func InitEventRouter() *EventRouter {
	res := &EventRouter{}

	return res
}

func (e *EventRouter) routing(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-e.incomingEventCh:
			message, err := typeConversion[incomingMessage](msg)
			if err != nil {
				log.Error(err)
				e.Shudown(err)
			}

			eventData, err := decodingMessage(message.GetMessage())
			if err != nil {
				log.Error(err)
				e.Shudown(err)
			}

			obj, ok := e.aggObjs[eventData.objectID]
			if !ok {
				// TODO: инициализировать новый объект
			}
			// отправка события объекту агрегации
			obj.eventReception(message.GetOffset(), eventData)

		}
	}

}

// метод прекращения работы EventRouter
func (e *EventRouter) Shudown(err error) {
	err = fmt.Errorf("%w: %w", stoppedEventRouterError{}, err)
	log.Error(err)
	e.cancel()
}

func (e *EventRouter) sendingEventToAggObj(offsetEvent int, event *eventData) {

}

func (e *EventRouter) getAggObj(objId int) *AggDataPerObject {
	obj, ok := e.aggObjs[objId]
	if !ok {

	}
}

func (e *EventRouter) createNewAggObj(objId int) *AggDataPerObject {

}
