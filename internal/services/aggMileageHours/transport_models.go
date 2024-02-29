package aggmileagehours

// события для отправки в горутину агрегации
type eventForAgg struct {
	offset    int64
	eventData *eventData
}

type mesForStorage struct {
	typeMes         string
	objectID        int
	shiftInitData   shiftObjData
	sessionInitData sessionDriverData
}

// метод для получения типа сообщения
func (m mesForStorage) GetType() string {
	return m.typeMes
}

// метод возвращает objID техники, для восстановления состояния
func (m mesForStorage) GetObjID() int {
	return m.objectID
}

// метод возвращает начальные данные для смены (данные для добавления новых данных в БД)
func (m mesForStorage) GetShiftData() interface{} {
	return m.shiftInitData

}

// метод возвращает начальные данные для сессии (данные для добавления новых данных в БД)
func (m mesForStorage) GetSessionData() interface{} {
	return m.sessionInitData
}

// транспортная структура (универсальный интерфейс)
type transportStruct struct {
	sender         string           // модуль отправитель сообщения
	mesage         mesForStorage    // сообщение отправителя
	reverseChannel chan interface{} // канал для отправки ответа от модуля storage
}

func (t transportStruct) GetSender() string {
	return t.sender
}

func (t transportStruct) GetMesage() interface{} {
	return t.mesage
}

// метод для отправки ответа от модуля storage
func (t transportStruct) SendAnswer(answer interface{}) {
	t.reverseChannel <- answer
}
