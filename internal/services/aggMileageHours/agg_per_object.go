package aggmileagehours

import (
	"context"
	"errors"
	"math"
	"time"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// структура в которой происходят процессы агрегации данных на объект техники
type AggDataPerObject struct {
	objectId                int                        // objectId техники
	shiftCurrentData        *ShiftObjData              // данные за текущую смену
	sessionCurrentData      *sessionDriverData         // данные текущей сессии водителя
	lastOffset              int64                      // последний offset загруженный в БД
	incomingCh              chan eventForAgg           // канал для получения событий
	storageCh               chan interface{}           // канал для связи с модулем работающем с БД
	cancel                  func()                     // функция для завршения конекста
	stateRestored           bool                       // флаг сигнализирующий и восстановленном состоянии, если флаг false занчит состояние еще не восстановлено
	settingsShift           *settingsDurationShifts    // настройки смены, меняются централизованно
	timeWaitResponseStorage int                        // время ожидания ответа от БД
	numAttemptRequest       int                        // количество попыток отправок запроса в модуль storage
	isActive                *activeFlag                // флаг активности обработчика
	timeMeter               *utils.ProcessingTimeMeter // измеритель времени процессов
}

func initAggDataPerObject(objectId, numAttemptRequest, timeWaitResponse int, settingsShift *settingsDurationShifts, storageCh chan interface{}, timeMeter *utils.ProcessingTimeMeter) (*AggDataPerObject, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	res := &AggDataPerObject{
		objectId:                objectId,
		incomingCh:              make(chan eventForAgg), // TODO: возможно нужен буферизированный канал, т.к. горутина может неуспеть обработать событие до отправки следующего
		cancel:                  cancel,
		settingsShift:           settingsShift,
		storageCh:               storageCh,
		timeWaitResponseStorage: timeWaitResponse,
		numAttemptRequest:       numAttemptRequest,
		isActive:                initActiveFlag(),
		timeMeter:               timeMeter,
	}

	// запуск горутины получения событий
	go res.gettingEvents(ctx)

	log.Infof("created aggDataObj with objectId: %d", objectId)
	return res, ctx
}

// процесс получения событий из маршрутизатора
func (a *AggDataPerObject) gettingEvents(ctx context.Context) {
	defer log.Warningf("gettingEvents for objectId: %d has finished", a.objectId)
	defer a.isActive.setIsActive(false)

	// восстановление состояния
	err := a.restoringState(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.incomingCh:
			// если offset событий меньше текущего offset тогда событие игнорируется
			if msg.offset > a.lastOffset {
				start := time.Now()
				err := a.eventHandling(ctx, msg.eventData, msg.offset)
				duration := time.Since(start)
				// отправка сообщения в измеритель
				a.timeMeter.SendMessToTimeMeter(utils.TrunsportToProcessingTime{
					NameProcess: "eventHandling",
					TimeProcess: duration,
				})
				if err != nil {
					log.Error(err)
					return
				}
			} else {
				log.Debugf("сообщение с offset: %d пропущено, текущий offset объекта: %d", msg.offset, a.lastOffset)
			}
		}
	}
}

// метод восстановления состояния
func (a *AggDataPerObject) restoringState(ctx context.Context) error {
	defer log.Warning("restoringState has been closed")
	log.Debugf("Восстановление состояния объекта: %d", a.objectId)

	mes := mesForStorage{
		typeMes:  restoreShiftDataPerObj,
		objectID: a.objectId,
	}

	// попытки отправки запросов в storage + обработка ошибок
	responseStorage, err := a.attemptSendRequest(ctx, mes)
	if err != nil {
		err = utils.Wrapper(restoringStateError{}, err)
		return err
	}

	err = responseStorage.GetError()
	if err != nil {
		switch err.Error() {
		case "error convert row to struct: no rows in result set":
			log.Debugf("нет данных для восстановления объекта %d", a.objectId)
			return nil
		default:
			// критическая ошибка
			err = utils.Wrapper(restoringStateError{}, err)
			return err
		}
	}

	// проверка ошибок конвертации
	shift, errDecodeShift := initNewShiftLoadingDBData(responseStorage.GetDataShift())
	session, errDecodeSession := initNewSessionLoadingDBData(responseStorage.GetDataDriverSession())
	if errDecodeShift != nil || errDecodeSession != nil {
		err = utils.Wrapper(restoringStateError{}, errDecodeShift)
		return err
	}

	a.shiftCurrentData = shift
	a.sessionCurrentData = session
	a.stateRestored = true
	log.Debugf("данные для объекта: %d восстановлены", a.objectId)
	a.lastOffset = int64(math.Min(float64(a.sessionCurrentData.offset), float64(a.sessionCurrentData.offset)))
	log.Debugf("offset для объекта: %d установлен: %d", a.objectId, a.lastOffset)

	return err
}

// метод обрабатывает событие:
func (a *AggDataPerObject) eventHandling(ctx context.Context, eventData *eventData, eventOffset int64) error {
	log.Debugf("Обработка сообщения для объекта: %d", a.objectId)
	var err error
	typeMes := updateShiftAndSession

	// получение номера и даты смены по времени сообщения, для проверки текущей смены объекта
	numShift, dateShift, err := a.settingsShift.defineShift(eventData.mesTime)
	if err != nil {
		return err
	}

	// создание смены, если она не была восстановлена (отсутствовали данные в БД)
	if !a.stateRestored {
		typeMes = addNewShiftAndSession
		a.shiftCurrentData = initNewShift(eventData, numShift, dateShift, eventOffset)
		a.sessionCurrentData = initNewSession(eventData, eventOffset)
		a.stateRestored = true
	} else {
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
	}
	// обновление локальных объектов сессии и смены
	a.updateObjects(eventData, eventOffset)

	// формирование сообщения для модуля storage
	mes := mesForStorage{
		typeMes:         typeMes,
		objectID:        a.objectId,
		shiftInitData:   *a.shiftCurrentData,
		sessionInitData: *a.sessionCurrentData,
	}

	// отправка сообщения в модуль storage (события будут обрабатываться по-разному, в зависимости от сообщения typeMes)
	// можно сделать отправку запросов через attemptSendRequest
	responceStorage, err := a.sendingMesToStorage(ctx, mes, a.timeWaitResponseStorage)
	if err != nil {
		return err
	}

	switch typeMes {
	case addNewShiftAndSession:
		responceData, err := сonversionAnswerStorage[responceStorageToAddNewShiftAndSession](responceStorage)
		if err != nil {
			// ошибка конвертации
			return err
		}
		if responceData.GetError() != nil {
			// ошибка в storage
			return err
		}
		log.Debugf("id новой смены: %d", responceData.GetShiftId())
		log.Debugf("id новой сессии: %d", responceData.GetSessionId())
		a.sessionCurrentData.setSessionId(responceData.GetSessionId())
		a.sessionCurrentData.setShiftId(responceData.GetShiftId())
		a.shiftCurrentData.setShiftId(responceData.GetShiftId())
		log.Debug("добавление новых записей смены и сессии успешно завершено")
	case updateShiftAndAddNewSession:
		// TODO: если другой интерфейс для получения ответа не нужно будет использовать то этот интерфейс надо переимоновать
		responceData, err := сonversionAnswerStorage[responceStorageToAddNewShiftAndSession](responceStorage)
		if err != nil {
			// ошибка конвертации
			return err
		}
		if responceData.GetError() != nil {
			// ошибка в storage
			return err
		}
		a.sessionCurrentData.setSessionId(responceData.GetSessionId())
		log.Debugf("данные смены обновлены, добавлена новая сессия с Id: %d", responceData.GetSessionId())
	case updateShiftAndSession:
		responceData, err := сonversionAnswerStorage[responceStorageToAddNewShiftAndSession](responceStorage)
		if err != nil {
			// ошибка конвертации
			return err
		}
		if responceData.GetError() != nil {
			// ошибка в storage
			return err
		}
		log.Debug("данные смены и данные сессии обновлены")
	}

	log.Debugf("оффсет объекта %d обновлен с %d на %d", a.objectId, a.lastOffset, eventOffset)
	a.lastOffset = eventOffset

	return err
}

// метод создания новых объектов (создается смена и сессия)
func (a *AggDataPerObject) createNewObjects(eventData *eventData, numShift int, dateShift time.Time) string {
	typeMes := addNewShiftAndSession
	// создается новый объект смены на основании данных старой смены
	a.shiftCurrentData = a.shiftCurrentData.createNewShift(numShift, dateShift, eventData)
	// создается новый объект сессии водителя на основании старой сессии
	a.sessionCurrentData = a.sessionCurrentData.createNewDriverSession(eventData)
	return typeMes
}

// метод создания сессии
func (a *AggDataPerObject) createSession(eventData *eventData) string {
	// тип сообщения которое будет сформировано для отправки в модуль storage
	typeMes := updateShiftAndAddNewSession
	a.sessionCurrentData = a.sessionCurrentData.createNewDriverSession(eventData)
	// установка id текущей смены для новой сессии
	a.sessionCurrentData.setShiftId(a.shiftCurrentData.Id)
	return typeMes
}

// метод обновляет объекты сессии и смены данными из событий
func (a *AggDataPerObject) updateObjects(eventData *eventData, eventOffset int64) {
	// обработка типа события (смена статуса загрузки)
	a.typeEventHandlig(eventData.typeEvent)
	// обновление объектов сессии и смены
	a.sessionCurrentData.updateSession(eventData, eventOffset, a.shiftCurrentData.Loaded)
	a.shiftCurrentData.updateShiftObjData(eventData, eventOffset, a.shiftCurrentData.Loaded)
}

// обертка над processAndSendToStorage, делает несколько попыток отправок запроса в модуль storage.
func (a *AggDataPerObject) attemptSendRequest(ctx context.Context, mes mesForStorage) (incomingMessageFromStorage, error) {
	var responseStorage incomingMessageFromStorage
	var err error

	for i := 0; i < a.numAttemptRequest; i++ {
		select {
		case <-ctx.Done():
			err = contextAggPerObjectClosedError{}
			return responseStorage, err
		default:
			responseStorage, err = a.processAndSendToStorage(ctx, mes)
			if err != nil {
				if !errors.Is(err, timeOutWaitAnswerDBError{}) {
					return responseStorage, err
				}
			} else {
				return responseStorage, err
			}
		}
	}
	err = attemptRequestError{}
	return responseStorage, err
}

// метод отправляет формирует сообщение и отправляет его в модуль storage,
// далее принимает ответ и конвертирует его в интерфейс ответа от storage:
//   - ctx: общий контекст обработчика событий
//   - mes: сообщение которое нужно отправить в storage
func (a *AggDataPerObject) processAndSendToStorage(ctx context.Context, mes mesForStorage) (incomingMessageFromStorage, error) {
	// нужно разделить обработку и отправку сообщения
	var err error
	var answerData incomingMessageFromStorage

	answer, err := a.sendingMesToStorage(ctx, mes, a.timeWaitResponseStorage)
	if err != nil {
		err = utils.Wrapper(processAndSendToStorageError{}, err)
		return answerData, err
	}

	answerData, err = сonversionAnswerStorage[incomingMessageFromStorage](answer)
	if err != nil {
		err = utils.Wrapper(processAndSendToStorageError{}, err)
	}

	return answerData, err
}

// метод отправляет сообщение в модуль storage и ожидает от него ответ, если ответ не успеет прийти за timeWait, то метод вернет ошибку
func (a *AggDataPerObject) sendingMesToStorage(ctx context.Context, mes mesForStorage, timeWait int) (interface{}, error) {
	defer log.Info("sendingMesToStorage закончил работу")
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
	log.Info("сообщение отправлено в storage")
	select {
	case <-ctx.Done():
		err = contextAggPerObjectClosedError{}
		log.Error(err)
		return answer, err
	case <-ctxTimeOut.Done():
		err = timeOutWaitAnswerDBError{}
		log.Error(err)
		return answer, err
	case answer := <-transportMes.reverseChannel:
		log.Info("принят ответ от storage")
		return answer, err
	}
}

// метод обработки типа события
func (a *AggDataPerObject) typeEventHandlig(typeEvent string) {
	switch typeEvent {
	case "DB_MSG_TYPE_LOAD":
		a.shiftCurrentData.Loaded = true

	case "DB_MSG_TYPE_UNLOAD":
		a.shiftCurrentData.Loaded = false
	}
}

// метод для отправки события в обработчик
func (a *AggDataPerObject) eventReception(offset int64, event *eventData) {
	a.incomingCh <- eventForAgg{
		offset:    offset,
		eventData: event,
	}
}

// метод для получения информации об активности обработчика
func (a *AggDataPerObject) getIsActive() bool {
	return a.isActive.getIsActive()
}
