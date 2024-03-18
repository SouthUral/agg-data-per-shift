package storage

// тип для отправки ответа модулю aggMileageHours
type responceForAggMileageHours struct {
	responceShiftSession[RowShiftObjData, RowSessionObjData]
}

// интерфейсный метод
func (r responceForAggMileageHours) GetDataShift() interface{} {
	return r.responseShift.data
}

func (r responceForAggMileageHours) GetErrorsResponceShift() (error, error) {
	return r.responseShift.criticalErr, r.responseShift.err
}

// интерфейсный метод
func (r responceForAggMileageHours) GetDataSession() interface{} {
	return r.responseSession.data
}

func (r responceForAggMileageHours) GetErrorsResponceSession() (error, error) {
	return r.responseSession.criticalErr, r.responseSession.err
}

// структура используется для ответа на запрос от модуля AggMileageHours на добавление новых записей в БД
type responceAggMileageHoursAddNewShiftAndSession struct {
	responceShiftSession[int, int]
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetDataShift() interface{} {
	return r.responseShift.data
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetErrorsResponceShift() (error, error) {
	return r.responseShift.criticalErr, r.responseShift.err
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetDataSession() interface{} {
	return r.responseSession.data
}

func (r responceAggMileageHoursAddNewShiftAndSession) GetErrorsResponceSession() (error, error) {
	return r.responseSession.criticalErr, r.responseSession.err
}

// структура возвращает ответ от метода производящего запрос в БД
type responceDataFromDB[D RowShiftObjData | RowSessionObjData | int] struct {
	data             D
	criticalErr, err error
}

// обновленная универсальная структура для ответа модулю AggMileageHours
type responceShiftSession[Shift RowShiftObjData | int, Session RowSessionObjData | int] struct {
	responseShift   responceDataFromDB[Shift]
	responseSession responceDataFromDB[Session]
	generalCriticalErr
}

type generalCriticalErr struct {
	criticalErr error
}

func (g generalCriticalErr) GetCriticalErr() error {
	return g.criticalErr
}
