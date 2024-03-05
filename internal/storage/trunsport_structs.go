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
