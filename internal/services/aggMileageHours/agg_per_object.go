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

// метод обрабатывает событие:
func (a *AggDataPerObject) eventHandling(ctx context.Context, eventData *eventData, eventOffset int64) error {
	var err error
	typeMes := updateShiftAndSession

	log.Debugf("%+v, mess_time: %s", eventData, eventData.mesTime)
	// получение номера и даты смены по времени сообщения, для проверки текущей смены объекта
	numShift, dateShift, err := a.settingsShift.defineShift(eventData.mesTime)
	if err != nil {
		return err
	}

	// если номер смены и дата смены не совпадают с номером и датой текущей смены, то нужно создать новую смену и сессию
	if !a.shiftCurrentData.checkDateNumCurrentShift(numShift, dateShift) {
		// получение номера и даты смены
		typeMes = a.createNewObjects(eventData, numShift, dateShift)
	}

	// если же дата и номер смены совпадают, далее нужно проверить не поменялась ли сессия водителя
	if !a.sessionCurrentData.checkDriverSession(eventData.numDriver) {
		// если id не совпадает с текущим то нужно создать новую сессию и обновить данные по смене
		typeMes = a.createSession(eventData)
	}

	// обновление локальных объектов сессии и смены
	a.updateObjects(eventData, eventOffset)

	// отправка сообщения в модуль storage (события будут обрабатываться по-разному, в зависимости от сообщения typeMes)
	answerFromStorage, err := a.processAndSendToStorage(ctx, typeMes)
	if err != nil {
		return err
	}

	switch typeMes {
	case addNewShiftAndSession:
		a.sessionCurrentData.setSessionId(answerFromStorage.driverSessionData.sessionId)
		a.sessionCurrentData.setShiftId(answerFromStorage.shiftData.id)
		a.shiftCurrentData.setShiftId(answerFromStorage.shiftData.id)
		log.Debug()
	case updateShiftAndAddNewSession:
		a.sessionCurrentData.setSessionId(answerFromStorage.driverSessionData.sessionId)
		log.Debug()
	case updateShiftAndSession:
		log.Debug()
	}

	return err
}

// метод создания новых объектов (создается смена и сессия)
func (a *AggDataPerObject) createNewObjects(eventData *eventData, numShift int, dateShift time.Time) string {
	typeMes := addNewShiftAndSession
	// создается новый объект смены на основании данных старой смены
	a.shiftCurrentData = a.shiftCurrentData.createNewShift(numShift, dateShift, eventData.mesTime)
	// создается новый объект сессии водителя на основании старой сессии
	a.sessionCurrentData = a.sessionCurrentData.createNewDriverSession(eventData.numDriver, eventData.mesTime)
	return typeMes
}

// метод создания сессии
func (a *AggDataPerObject) createSession(eventData *eventData) string {
	// тип сообщения которое будет сформировано для отправки в модуль storage
	typeMes := updateShiftAndAddNewSession
	a.sessionCurrentData = a.sessionCurrentData.createNewDriverSession(eventData.numDriver, eventData.mesTime)
	// установка id текущей смены для новой сессии
	a.sessionCurrentData.setShiftId(a.shiftCurrentData.id)
	return typeMes
}

// метод обновляет объекты сессии и смены данными из событий
func (a *AggDataPerObject) updateObjects(eventData *eventData, eventOffset int64) {
	// обработка типа события (смена статуса загрузки)
	a.typeEventHandlig(eventData.typeEvent)
	// обновление объектов сессии и смены
	a.sessionCurrentData.updateSession(eventData, eventOffset, a.shiftCurrentData.loaded)
	a.shiftCurrentData.updateShiftObjData(eventData, eventOffset, a.shiftCurrentData.loaded)
}

// метод отправляет формирует сообщение и отправляет его в модуль storage,
// далее принимает ответ и конвертирует его во внутренние интерфейсы
func (a *AggDataPerObject) processAndSendToStorage(ctx context.Context, typeMes string) (storageAnswerData, error) {
	var err error
	var answerData storageAnswerData

	mesForStorage := mesForStorage{
		typeMes:         typeMes,
		objectID:        a.objectId,
		shiftInitData:   *a.shiftCurrentData,
		sessionInitData: *a.sessionCurrentData,
	}

	answer, err := a.sendingMesToStorage(ctx, mesForStorage, a.timeWaitResponseStorage)
	if err != nil {
		err = fmt.Errorf("%w: %w", processAndSendToStorageError{}, err)
		return answerData, err
	}

	answerData, err = сonversionAnswerStorage(answer)
	if err != nil {
		err = fmt.Errorf("%w: %w", processAndSendToStorageError{}, err)
		return answerData, err
	}

	return answerData, err
}

// метод отправляет сообщение в модуль storage и ожидает от него ответ, если ответ не успеет прийти за timeWait, то метод вернет ошибку
func (a *AggDataPerObject) sendingMesToStorage(ctx context.Context, mes mesForStorage, timeWait int) (interface{}, error) {
	var answer interface{}
	var err error

	ctxTimeOut, _ := context.WithTimeout(context.Background(), time.Duration(timeWait)*time.Second)
	reverseChannel := make(chan interface{})

	transportMes := transportStruct{
		sender:         nameSender,
		mesage:         mes,
		reverseChannel: reverseChannel,
	}

	a.storageCh <- transportMes
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

// метод для отправки события в обработчик
func (a *AggDataPerObject) eventReception(offset int64, event *eventData) {
	a.incomingCh <- eventForAgg{
		offset:    offset,
		eventData: event,
	}
}
