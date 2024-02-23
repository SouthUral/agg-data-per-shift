package aggmileagehours

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// структура в которой происходят процессы агрегации данных на объект техники
type AggDataPerObject struct {
	objectId                int                     // objectId техники
	shiftData               shiftObjData            // данные за смену
	lastOffset              int64                   // последний offset загруженный в БД
	incomingCh              chan eventForAgg        // канал для получения событий
	storageCh               chan interface{}        // канал для связи с модулем работающем с БД
	cancel                  func()                  // функция для завршения конекста
	stateRestored           bool                    // флаг сигнализирующий и восстановленном состоянии, если флаг false занчит состояние еще не восстановлено
	settingsShift           *settingsDurationShifts // настройки смены, меняются централизованно
	timeWaitResponseStorage int                     // время ожидания ответа от БД
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
				// если состояние не восстановлено (объект был только что создан) то восстанавливаем состояние
				// на время восстановления горутина блокируется, сообщения собираются в канале
				if !a.stateRestored {
					a.restoringState(ctx)
				}
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
	log.Debugf("%+v, mess_time: %s", eventData, eventData.mesTime)

}

// метод для отправки события в обработчик
func (a *AggDataPerObject) eventReception(offset int64, event *eventData) {
	a.incomingCh <- eventForAgg{
		offset:    offset,
		eventData: event,
	}
}

// метод восстановления состояния
func (a *AggDataPerObject) restoringState(ctx context.Context) error {
	// TODO: нужно отправить в sotrage запрос на восстановление состояния
	// Данные, которые будут переданы в запросе это objectID техники
	// Данные, которые должны быть получены:
	// - строка из таблицы Смены
	// - строка из таблицы Сессий водителей
	// если данных не будет, то далее будут созданы новые сессия и смена
	mes := mesForStorage{
		typeMes:  restoreShiftDataPerObj,
		objectID: a.objectId,
	}

	answer, err := a.sendingMesToStorage(ctx, mes, a.timeWaitResponseStorage)
	// TODO: нужна обработка ошибки
	if err != nil {
		return err
	}

	answerData, err := a.typeConversionAnswerStorage(answer)
	if err != nil {
		return err
	}
	// TODO: теперь нужно перенести все данные из интерфейсов во внутренние структуры сессии и смены

}

// метод создания новых объектов (создается смена и сессия)
func (a *AggDataPerObject) createNewObjects() {
	// получение  id смены и id сессии

}

// метод создания сессии
func (a *AggDataPerObject) createSession() {
	// получение id сессии
}

// обновление объектов (смены и сессии)
func (a *AggDataPerObject) updateObjects() {
	// метод не возвращает никаких данных
}

// метод отправляет сообщение в модуль storage и ожидает от него ответ, если ответ не успеет прийти за timeWait, то метод вернет ошибку
func (a *AggDataPerObject) sendingMesToStorage(ctx context.Context, mes mesForStorage, timeWait int) (interface{}, error) {
	var answer interface{}
	var err error

	ctxTimeOut, _ := context.WithTimeout(context.Background(), time.Duration(timeWait)*time.Second)
	reverseChannel := make(chan interface{})
	mes.reverseChannel = reverseChannel
	a.storageCh <- mes
	select {
	case <-ctx.Done():
		err = contextAggPerObjectClosedError{}
		return answer, err
	case <-ctxTimeOut.Done():
		err = timeOutWaitAnswerDBError{}
		return answer, err
	case answer := <-reverseChannel:
		return answer, err
	}
}

// метод для преобразования ответа от модуля storage
func (a *AggDataPerObject) typeConversionAnswerStorage(answer interface{}) (storageAnswerData, error) {
	var err error
	convertedStorageData := storageAnswerData{}

	storageAnswer, err := typeConversion[incomingMessageFromStorage](answer)
	if err != nil {
		err = fmt.Errorf("%w: %w", typeConversionAnswerStorageDataError{}, err)
		return convertedStorageData, err
	}

	dataShift, err := typeConversion[dataShiftFromStorage](storageAnswer.GetDataShift())
	if err != nil {
		err = fmt.Errorf("%w: %w", typeConversionAnswerStorageDataError{}, err)
		return convertedStorageData, err
	}
	convertedStorageData.shiftData = dataShift

	dataDriverSession, err := typeConversion[dataDriverSessionFromStorage](storageAnswer.GetDataDriverSession())
	if err != nil {
		err = fmt.Errorf("%w: %w", typeConversionAnswerStorageDataError{}, err)
		return convertedStorageData, err
	}

	convertedStorageData.driverSessionData = dataDriverSession

	return convertedStorageData, err
}
