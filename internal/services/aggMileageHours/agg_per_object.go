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
	shiftCurrentData        *shiftObjData           // данные за текущую смену
	sessionCurrentData      *sessionDriverData      // данные текущей сессии водителя
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
					err := a.restoringState(ctx)
					log.Error(err)
					return
				}
				err := a.eventHandling(ctx, msg.eventData, msg.offset)
				if err != nil {
					log.Error(err)
					return
				}
			}
		}
	}
}

// метод обрабатывает событие:
func (a *AggDataPerObject) eventHandling(ctx context.Context, eventData *eventData, eventOffset int64) error {
	var err error
	// TODO: работа с событиями
	// - нужно определить текущее сообщение относится к полученной смене
	// если данные события относятся к полученной смене/сессии то записи в БД (по id) будут обновлятсья
	// - проверка смены (к какой смене относится события)
	// - проверка сессии (проверка, тот ли водитель сейчас)
	// - для определения смены нужно сделать запрос к settingsShift (передать туда дату, время события)
	// - на выходе должны образоваться дата смены и ее номер (если они совпадают с датой смены и номером то это та смена)
	// - если на выходе получится другая дата или номер смены то нужно создать новый объект смены и сессии
	// - если проверка смены прошла (смена не поменялась) то следущая идет проверка сессии (просто сравнить id водителя)
	log.Debugf("%+v, mess_time: %s", eventData, eventData.mesTime)
	// получение номера и даты смены по времени сообщения
	numShift, dateShift, err := a.settingsShift.defineShift(eventData.mesTime)
	if err != nil {
		return err
	}

	if !a.shiftCurrentData.checkDateNumCurrentShift(numShift, dateShift) {
		// если номер смены и дата смены не совпадают с номером и датой текущей смены, то нужно создать новую смену и сессию
		err = a.createNewObjects(ctx, eventData, eventOffset)
		if err != nil {
			return err
		}
	}

	// если же дата и номер смены совпадают, далее нужно проверить не поменялась ли сессия водителя
	if !a.sessionCurrentData.checkDriverSession(eventData.numDriver) {
		// если id не совпадает с текущим то нужно создать новую сессию и обновить данные по смене
		//
		err = a.createSession(ctx, eventData)
		if err != nil {
			return err
		}
	}

	// если все проверки пройдены смена и сессия остались теми же, нужно обновить данные в локальных структурах и в БД
	err = a.updateObjects(ctx, eventData)

	return err
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
	// TODO: примечание, если в БД не было записей, то нужно сгенерировать новую смену и сессию
	mes := mesForStorage{
		typeMes:  restoreShiftDataPerObj,
		objectID: a.objectId,
	}

	answer, err := a.sendingMesToStorage(ctx, mes, a.timeWaitResponseStorage)
	if err != nil {
		err = fmt.Errorf("%w: %w", restoringStateError{}, err)
		return err
	}

	answerData, err := сonversionAnswerStorage(answer)
	if err != nil {
		err = fmt.Errorf("%w: %w", restoringStateError{}, err)
		return err
	}

	a.loadingStorageData(answerData)
	a.stateRestored = true

	return err
}

// метод создания новых объектов (создается смена и сессия)
func (a *AggDataPerObject) createNewObjects(ctx context.Context, eventData *eventData, eventOffset int64) error {
	var err error
	// получение  id смены и id сессии из БД
	numShift, dateShift, err := a.settingsShift.defineShift(eventData.mesTime)
	if err != nil {
		return err
	}

	// создается новый объект смены на основании данных старой смены
	a.shiftCurrentData = a.shiftCurrentData.createNewShift(numShift, dateShift, eventData.mesTime)
	// создается новый объект сессии водителя на основании старой сессии
	a.sessionCurrentData = a.sessionCurrentData.createNewDriverSession(eventData.numDriver, eventData.mesTime)
	// если предыдущая смена закончилась с загруженным транспортом то нужно посмотреть, не пришло ли в текущем событии событие разгрузка
	// так же первым сообщением может быть начало погрузки
	a.typeEventHandlig(eventData.typeEvent)

	// обновление созданных объектов
	a.sessionCurrentData.updateSession(eventData, eventOffset, a.shiftCurrentData.loaded)
	a.shiftCurrentData.updateShiftObjData(eventData, eventOffset, a.shiftCurrentData.loaded)

	// отправить сообщение в модуль storage с данными новой смены и сессии
	mesForStorage := mesForStorage{
		typeMes:         addNewShiftAndSession,
		objectID:        a.objectId,
		shiftInitData:   initShiftObjTransportData(*a.shiftCurrentData),
		sessionInitData: initSessionDriverTransportData(*a.sessionCurrentData),
	}

	answer, err := a.sendingMesToStorage(ctx, mesForStorage, a.timeWaitResponseStorage)
	// TODO: нужно обработать ответ от модуля storage, нужны id смены и сессии
	return err
}

// метод создания сессии
func (a *AggDataPerObject) createSession(ctx context.Context, eventData *eventData) error {
	var err error
	// получение id сессии
	return err
}

// обновление объектов (смены и сессии)
func (a *AggDataPerObject) updateObjects(ctx context.Context, eventData *eventData) error {
	var err error
	// метод не возвращает никаких данных
	return err
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

// метод загружает данные полученные из storage интерфейсов в локальные структуры
func (a *AggDataPerObject) loadingStorageData(data storageAnswerData) {
	a.sessionCurrentData = data.driverSessionData
	a.shiftCurrentData = data.shiftData
}

// метод обработки типа события
func (a *AggDataPerObject) typeEventHandlig(typeEvent string) {
	switch typeEvent {
	case "DB_MSG_TYPE_START_LOAD":
		a.shiftCurrentData.loaded = true

	case "DB_MSG_TYPE_UNLOAD":
		a.shiftCurrentData.loaded = false
	}
}
