package storage

// тип для отправки ответа модулю aggMileageHours
type answerForAggMileageHours struct {
	shiftData   RowShiftObjData
	sessionData RowSessionObjData
	err         error
}

// интерфейсный метод
func (a answerForAggMileageHours) GetDataShift() interface{} {
	return a.shiftData
}

// интерфейсный метод
func (a answerForAggMileageHours) GetDataDriverSession() interface{} {
	return a.sessionData
}

// интерфейсный метод
func (a answerForAggMileageHours) GetError() error {
	return a.err
}

// структура используется для ответа на запрос от модуля AggMileageHours на добавление новых записей в БД
type responceAggMileageHoursAddNewShiftAndSession struct {
	shiftId   int
	sessionId int
	err       error
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetShiftId() int {
	return r.shiftId
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetSessionId() int {
	return r.sessionId
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetError() error {
	return r.err
}
