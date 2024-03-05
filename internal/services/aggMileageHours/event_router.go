package aggmileagehours

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	storage "agg-data-per-shift/internal/storage"
	utils "agg-data-per-shift/pkg/utils"
)

// структура в которой происходит распределение событий по горутинам объектов техники
// ---
type EventRouter struct {
	incomingEventCh  chan interface{}          // канал для получения событий
	storageCh        chan interface{}          // канал для связи с модулем psql
	aggObjs          map[int]*AggDataPerObject // map с объектами агрегации данных (по id техники)
	settingShift     *settingsDurationShifts   // настройки смены
	pgConn           *storage.PgConn
	cancel           func()
	timeWaitResponse int // для горутин обработчиков, время ожидания ответа от БД
}

func InitEventRouter(storageCh chan interface{}, timeWaitResponse int) (*EventRouter, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	res := &EventRouter{
		incomingEventCh:  make(chan interface{}),
		cancel:           cancel,
		aggObjs:          make(map[int]*AggDataPerObject),
		storageCh:        storageCh,
		timeWaitResponse: timeWaitResponse,
	}

	go res.routing(ctx)

	return res, ctx
}

func (e *EventRouter) GetIncomingEventCh() chan interface{} {
	return e.incomingEventCh
}

func (e *EventRouter) routing(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-e.incomingEventCh:
			message, err := utils.TypeConversion[incomingMessage](msg)
			if err != nil {
				log.Error(err)
				e.Shudown(err)
				return
			}

			eventData, err := decodingMessage(message.GetMsg())
			if err != nil {
				log.Error(err)
				e.Shudown(err)
				return
			}

			e.sendingEventToAggObj(message.GetOffset(), eventData)
		}
	}

}

// метод прекращения работы EventRouter
func (e *EventRouter) Shudown(err error) {
	err = fmt.Errorf("%w: %w", stoppedEventRouterError{}, err)
	log.Error(err)
	e.cancel()
}

func (e *EventRouter) sendingEventToAggObj(offsetEvent int64, event *eventData) {
	obj := e.getAggObj(event.objectID)
	obj.eventReception(offsetEvent, event)
}

func (e *EventRouter) getAggObj(objId int) *AggDataPerObject {
	obj, ok := e.aggObjs[objId]
	if !ok {
		obj = e.createNewAggObj(objId)
	}
	return obj
}

func (e *EventRouter) createNewAggObj(objId int) *AggDataPerObject {
	aggObj, _ := initAggDataPerObject(objId, 100, e.timeWaitResponse, e.settingShift, e.storageCh)
	e.aggObjs[objId] = aggObj
	return aggObj
}

// метод для отрправки событий в роутер
func (e *EventRouter) EventReception(event interface{}) {
	e.incomingEventCh <- event
}
