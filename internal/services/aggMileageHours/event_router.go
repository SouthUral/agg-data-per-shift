package aggmileagehours

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	utils "agg-data-per-shift/pkg/utils"
)

// структура в которой происходит распределение событий по горутинам объектов техники
// ---
type EventRouter struct {
	incomingEventCh chan interface{}           // канал для получения событий
	storageCh       chan interface{}           // канал для связи с модулем psql
	aggObjs         map[int]*AggDataPerObject  // map с объектами агрегации данных (по id техники)
	settingShift    *settingsDurationShifts    // настройки смены
	timeMeter       *utils.ProcessingTimeMeter // измеритель времени процессов
	cancel          func()
	activeFlag      *activeFlag
}

func InitEventRouter(storageCh chan interface{}, timeMeter *utils.ProcessingTimeMeter) (*EventRouter, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	res := &EventRouter{
		incomingEventCh: make(chan interface{}, 1000),
		cancel:          cancel,
		aggObjs:         make(map[int]*AggDataPerObject),
		storageCh:       storageCh,
		settingShift:    initSettingsDurationShifts(-4),
		timeMeter:       timeMeter,
		activeFlag:      initActiveFlag(),
	}

	// временно сам добавляю смены
	t1, _ := time.Parse(time.TimeOnly, "00:00:00")
	t2, _ := time.Parse(time.TimeOnly, "12:00:00")
	res.settingShift.AddShiftSetting(1, 12, t1)
	res.settingShift.AddShiftSetting(2, 12, t2)

	go res.routing(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if !res.activeFlag.getIsActive() {
					res.Shudown(errHadlerEventError)
					log.Debug("ПРОВЕРКА ФЛАГА ВЫЯВИЛА ОШИБКУ")
					return
				}
				log.Debugf("ЗАПОЛЕННОСТЬ БУФЕРА МАРШРУТИЗАТОРА %d", len(res.incomingEventCh))
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return res, ctx
}

func (e *EventRouter) GetIncomingEventCh() chan interface{} {
	return e.incomingEventCh
}

func (e *EventRouter) routing(ctx context.Context) {
	defer log.Debug("РОУТЕР ПРЕКРАТИЛ РАБОТУ!")
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
			log.Debugf("получен Offset: %d", message.GetOffset())
			eventData, err := decodingMessage(message.GetMsg())
			if err != nil {
				log.Error(err)
				e.Shudown(err)
				return
			}

			e.sendingEventToAggObj(ctx, message.GetOffset(), eventData)
		}
	}

}

// метод прекращения работы EventRouter
func (e *EventRouter) Shudown(err error) {
	err = fmt.Errorf("%w: %w", stoppedEventRouterError{}, err)
	log.Error(err)
	e.cancel()
	for _, obj := range e.aggObjs {
		obj.Shudown()
	}
}

// поиск объекта обработки, и отправка сообщения
func (e *EventRouter) sendingEventToAggObj(ctx context.Context, offsetEvent int64, event *eventData) {
	obj := e.getAggObj(event.objectID)
	obj.eventReception(ctx, offsetEvent, event)
}

func (e *EventRouter) getAggObj(objId int) *AggDataPerObject {
	obj, ok := e.aggObjs[objId]
	if !ok {
		obj = e.createNewAggObj(objId)
	}
	return obj
}

func (e *EventRouter) createNewAggObj(objId int) *AggDataPerObject {
	aggObj, _ := initAggDataPerObject(objId, 100, e.settingShift, e.storageCh, e.timeMeter, e.activeFlag)
	e.aggObjs[objId] = aggObj
	return aggObj
}

// метод для отрправки событий в роутер
func (e *EventRouter) EventReception(ctx context.Context, event interface{}) error {
	if !e.activeFlag.getIsActive() {
		return errActiveEventRouter
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case e.incomingEventCh <- event:
			return nil
		default:
			if !e.activeFlag.getIsActive() {
				return errActiveEventRouter
			}
		}
	}
}
