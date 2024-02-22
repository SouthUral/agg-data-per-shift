package aggmileagehours

// события для отправки в горутину агрегации
type eventForAgg struct {
	offset    int64
	eventData *eventData
}

type mesForStorage struct {
	typeMes        string
	objectID       int
	reverseChannel chan interface{} // канал для отправки ответа от модуля storage
	// TODO: данные которые могут быть переданы:
	// - тип сообщения
	// - objID техники, для восстановления состояния
	// - начальные данные для смены
	// - начальные данные для сессии
	// - обновление данных, передается только id смены и id сессии
}

// метод для отправки ответа от модуля storage
func (m mesForStorage) SendAnswer(answer interface{}) {
	m.reverseChannel <- answer
}

// метод для получения типа сообщения
func (m mesForStorage) GetType() string {
	return m.typeMes
}

// метод возвращает objID техники, для восстановления состояния
func (m mesForStorage) GetObjID() int {
	return m.objectID
}

// метод возвращает начальные данные для смены
func (m mesForStorage) GetInitialShiftData() {

}

// метод возвращает начальные данные для сессии
func (m mesForStorage) GetInitialSessionData() {

}

// метод для получения данных на обновление смены
func (m mesForStorage) GetUpdateShiftData() {

}

// метод для получения данных на обновлении сессии
func (m mesForStorage) GetUpdateSessionData() {

}
