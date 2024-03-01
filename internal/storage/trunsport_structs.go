package storage

// тип для отправки ответа модулю aggMileageHours
type answerForAggMileageHours struct {
	shiftData   []byte
	sessionData []byte
	err         error
}

// интерфейсный метод
func (a answerForAggMileageHours) GetDataShift() []byte {
	return a.shiftData
}

// интерфейсный метод
func (a answerForAggMileageHours) GetDataDriverSession() []byte {
	return a.sessionData
}

// интерфейсный метод
func (a answerForAggMileageHours) GetError() error {
	return a.err
}
