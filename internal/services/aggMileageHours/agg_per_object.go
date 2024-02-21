package aggmileagehours

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// структура в которой происходят процессы агрегации данных на объект техники
type AggDataPerObject struct {
	objectId      int                     // objectId техники
	shiftData     shiftObjData            // данные за смену
	lastOffset    int64                   // последний offset загруженный в БД
	incomingCh    chan eventForAgg        // канал для получения событий
	cancel        func()                  // функция для завршения конекста
	settingsShift *settingsDurationShifts // настройки смены, меняются централизованно
}

// TODO: параметры смены будут в отдельном модуле, который будет отправлять информацию в случае изменения, горутины работают со своими локальными настройками
// либо можно оставить общий объект настроек смены но в роутере, в самом роутере обновлять настройки при получении от модуля

// TODO: при инициализации нужно полностью восстановить информацию о смене
// TODO: горутина сама восстанавливает информацию о смене, после создания она отправляет запрос в БД на восстановление состояния

func initAggDataPerObject(objectId int, settingsShift *settingsDurationShifts) (*AggDataPerObject, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	res := &AggDataPerObject{
		objectId:      objectId,
		incomingCh:    make(chan eventForAgg), // TODO: возможно нужен буферизированный канал, т.к. горутина может неуспеть обработать событие до отправки следующего
		cancel:        cancel,
		settingsShift: settingsShift,
	}

	// при запуске нужно восстановить состояние, нужно отправить запрос в БД на восстановление состояния
	// данные о текущем состоянии сохраняются в AggDataPerObject.shiftData
	// запуск горутины чтения должен происходить после восстановления состояния

	// запуск горутины получения событий
	go res.gettingEvents(ctx)

	log.Infof("created aggDataObj with objectId: %d", objectId)
	return res, ctx
}

// процесс получения событий из маршрутизатора
func (a *AggDataPerObject) gettingEvents(ctx context.Context) {
	defer log.Warningf("gettingEvents for objectId: %d has finished", a.objectId)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.incomingCh:
			// если offset событий меньше текущего offset тогда событие игнорируется
			if msg.offset > a.lastOffset {
				a.eventHandling(msg.eventData)
			}
		}
	}
}

// метод обрабатывает событие:
func (a *AggDataPerObject) eventHandling(eventData *eventData) {
	// - TODO: нужно определить в к какой смене относится событие;
	// - TODO: нужно определить, не поменялся ли водитель на технике;
	// - TODO:
	log.Debugf("%v", eventData)

}

// метод для отправки события в обработчик
func (a *AggDataPerObject) eventReception(offset int64, event *eventData) {
	a.incomingCh <- eventForAgg{
		offset:    offset,
		eventData: event,
	}
}
