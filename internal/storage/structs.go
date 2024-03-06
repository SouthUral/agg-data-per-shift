package storage

import (
	"fmt"
	"time"

	utils "agg-data-per-shift/pkg/utils"
)

type EngHoursObjData struct {
	EngHoursStart   float32 `db:"eng_hours_start" json:"eng_hours_start"`     // моточасы на начало
	EngHoursCurrent float32 `db:"eng_hours_current" json:"eng_hours_current"` // последняя обновленноая запись моточасов
	EngHoursEnd     float32 `db:"eng_hours_end" json:"eng_hours_end"`         // моточасы на конец
}

func (e EngHoursObjData) GetEngHoursStart() float32 {
	return e.EngHoursStart
}
func (e EngHoursObjData) GetEngHoursCurrent() float32 {
	return e.EngHoursCurrent
}
func (e EngHoursObjData) GetEngHoursEnd() float32 {
	return e.EngHoursEnd
}

type MileageObjData struct {
	MileageStart                int `db:"mileage_start" json:"mileage_start"`                                     // пробег на начало (смены/сессии)
	MileageCurrent              int `db:"mileage_current" json:"mileage_current"`                                 // текущий пробег
	MileageEnd                  int `db:"mileage_end" json:"mileage_end"`                                         // проебег на конец смены
	MileageLoaded               int `db:"mileage_loaded" json:"mileage_loaded"`                                   // пробег в груженом состоянии
	MileageAtBeginningOfLoading int `db:"mileage_at_beginning_of_loading" json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageEmpty                int `db:"mileage_empty" json:"mileage_empty"`                                     // пробег в порожнем состоянии
}

func (m MileageObjData) GetMileageStart() int {
	return m.MileageStart
}
func (m MileageObjData) GetMileageCurrent() int {
	return m.MileageCurrent
}
func (m MileageObjData) GetMileageEnd() int {
	return m.MileageEnd
}
func (m MileageObjData) GetMileageLoaded() int {
	return m.MileageLoaded
}
func (m MileageObjData) GetMileageAtBeginningOfLoading() int {
	return m.MileageAtBeginningOfLoading
}
func (m MileageObjData) GetMileageEmpty() int {
	return m.MileageEmpty
}

type MileageObjGPSData struct {
	MileageGPSStart                int `db:"mileage_gps_start" json:"mileage_gps_start"`                                     // пробег на начало (смены/сессии)
	MileageGPSCurrent              int `db:"mileage_gps_current" json:"mileage_gps_current"`                                 // текущий пробег
	MileageGPSEnd                  int `db:"mileage_gps_end" json:"mileage_gps_end"`                                         // проебег на конец смены
	MileageGPSLoaded               int `db:"mileage_gps_loaded" json:"mileage_gps_loaded"`                                   // пробег в груженом состоянии
	MileageGPSAtBeginningOfLoading int `db:"mileage_gps_at_beginning_of_loading" json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageGPSEmpty                int `db:"mileage_gps_empty" json:"mileage_gps_empty"`
}

func (m MileageObjGPSData) GetMileageStart() int {
	return m.MileageGPSStart
}
func (m MileageObjGPSData) GetMileageCurrent() int {
	return m.MileageGPSCurrent
}
func (m MileageObjGPSData) GetMileageEnd() int {
	return m.MileageGPSEnd
}
func (m MileageObjGPSData) GetMileageLoaded() int {
	return m.MileageGPSLoaded
}
func (m MileageObjGPSData) GetMileageAtBeginningOfLoading() int {
	return m.MileageGPSAtBeginningOfLoading
}
func (m MileageObjGPSData) GetMileageEmpty() int {
	return m.MileageGPSEmpty
}

// данные смены (для получения из БД)
type RowShiftObjData struct {
	Id              int       `db:"id" json:"id"`                               // id текущей смены
	NumShift        int       `db:"num_shift" json:"num_shift"`                 // номер смены (1/2....n)
	ShiftDateStart  time.Time `db:"shift_date_start" json:"shift_date_start"`   // время и дата начала смены (время первого события смены)
	ShiftDateEnd    time.Time `db:"shift_date_end" json:"shift_date_end"`       // время окончания смены (время последнего обновления)
	ShiftDate       time.Time `db:"shift_date" json:"shift_date"`               // текущая дата смены (смена может начинаться в другую дату)
	UpdatedTime     time.Time `db:"updated_time" json:"updated_time"`           // время обновления смены
	Offset          int64     `db:"event_offset" json:"event_offset"`           // offset события, которое последнее обновило состояние смены
	CurrentDriverId int       `db:"current_driver_id" json:"current_driver_id"` // id текущего водителя техники (может меняться в пределах смены)
	Loaded          bool      `db:"loaded" json:"loaded"`                       // флаг, техника на момент обновления записи находилась в груженом состоянии
	engHours        EngHoursObjData
	mileageData     MileageObjData
	mileageGPSData  MileageObjGPSData

	// EngHoursStart   float32 `db:"eng_hours_start" json:"eng_hours_start"`     // моточасы на начало
	// EngHoursCurrent float32 `db:"eng_hours_current" json:"eng_hours_current"` // последняя обновленноая запись моточасов
	// EngHoursEnd     float32 `db:"eng_hours_end" json:"eng_hours_end"`         // моточасы на конец

	// MileageStart                int `db:"mileage_start" json:"mileage_start"`                                     // пробег на начало (смены/сессии)
	// MileageCurrent              int `db:"mileage_current" json:"mileage_current"`                                 // текущий пробег
	// MileageEnd                  int `db:"mileage_end" json:"mileage_end"`                                         // проебег на конец смены
	// MileageLoaded               int `db:"mileage_loaded" json:"mileage_loaded"`                                   // пробег в груженом состоянии
	// MileageAtBeginningOfLoading int `db:"mileage_at_beginning_of_loading" json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	// MileageEmpty                int `db:"mileage_empty" json:"mileage_empty"`                                     // пробег в порожнем состоянии

	// MileageGPSStart                int `db:"mileage_gps_start" json:"mileage_gps_start"`                                     // пробег на начало (смены/сессии)
	// MileageGPSCurrent              int `db:"mileage_gps_current" json:"mileage_gps_current"`                                 // текущий пробег
	// MileageGPSEnd                  int `db:"mileage_gps_end" json:"mileage_gps_end"`                                         // проебег на конец смены
	// MileageGPSLoaded               int `db:"mileage_gps_loaded" json:"mileage_gps_loaded"`                                   // пробег в груженом состоянии
	// MileageGPSAtBeginningOfLoading int `db:"mileage_gps_at_beginning_of_loading" json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	// MileageGPSEmpty                int `db:"mileage_gps_empty" json:"mileage_gps_empty"`                                     // пробег в порожнем состоянии
}

func (r RowShiftObjData) GetShiftId() int {
	return r.Id
}
func (r RowShiftObjData) GetShiftNum() int {
	return r.NumShift
}
func (r RowShiftObjData) GetShiftDateStart() time.Time {
	return r.ShiftDateStart
}
func (r RowShiftObjData) GetShiftDateEnd() time.Time {
	return r.ShiftDateEnd
}
func (r RowShiftObjData) GetShiftDate() time.Time {
	return r.ShiftDate
}
func (r RowShiftObjData) GetUpdatedTime() time.Time {
	return r.UpdatedTime
}
func (r RowShiftObjData) GetOffset() int64 {
	return r.Offset
}
func (r RowShiftObjData) GetCurrentDriverId() int {
	return r.CurrentDriverId
}
func (r RowShiftObjData) GetStatusLoaded() bool {
	return r.Loaded
}
func (r RowShiftObjData) GetEngHoursData() interface{} {
	return r.engHours
}
func (r RowShiftObjData) GetMileageData() interface{} {
	return r.mileageData
}
func (r RowShiftObjData) GetMileageGPSData() interface{} {
	return r.mileageGPSData
}

// данные сессии (для получения из БД)
type RowSessionObjData struct {
	ShiftId           int       `db:"shift_id" json:"shift_id"`                       // id смены, в которой находится сессия
	SessionId         int       `db:"id" json:"id"`                                   // id сессии, берется из БД
	DriverId          int       `db:"driver_id" json:"driver_id"`                     // id водителя
	Offset            int64     `db:"event_offset" json:"event_offset"`               // последний записанный offset
	TimeStartSession  time.Time `db:"time_start_session" json:"time_start_session"`   // время старта сессии
	TimeUpdateSession time.Time `db:"time_update_session" json:"time_update_session"` // время обновления записи сессии
	AvSpeed           float32   `db:"av_speed" json:"av_speed"`                       // средняя скорость водителя
	engHours          EngHoursObjData
	mileageData       MileageObjData
	mileageGPSData    MileageObjGPSData

	// EngHoursStart   float32 `db:"eng_hours_start" json:"eng_hours_start"`     // моточасы на начало
	// EngHoursCurrent float32 `db:"eng_hours_current" json:"eng_hours_current"` // последняя обновленноая запись моточасов
	// EngHoursEnd     float32 `db:"eng_hours_end" json:"eng_hours_end"`         // моточасы на конец

	// MileageStart                int `db:"mileage_start" json:"mileage_start"`                                     // пробег на начало (смены/сессии)
	// MileageCurrent              int `db:"mileage_current" json:"mileage_current"`                                 // текущий пробег
	// MileageEnd                  int `db:"mileage_end" json:"mileage_end"`                                         // проебег на конец смены
	// MileageLoaded               int `db:"mileage_loaded" json:"mileage_loaded"`                                   // пробег в груженом состоянии
	// MileageAtBeginningOfLoading int `db:"mileage_at_beginning_of_loading" json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	// MileageEmpty                int `db:"mileage_empty" json:"mileage_empty"`                                     // пробег в порожнем состоянии

	// MileageGPSStart                int `db:"mileage_gps_start" json:"mileage_gps_start"`                                     // пробег на начало (смены/сессии)
	// MileageGPSCurrent              int `db:"mileage_gps_current" json:"mileage_gps_current"`                                 // текущий пробег
	// MileageGPSEnd                  int `db:"mileage_gps_end" json:"mileage_gps_end"`                                         // проебег на конец смены
	// MileageGPSLoaded               int `db:"mileage_gps_loaded" json:"mileage_gps_loaded"`                                   // пробег в груженом состоянии
	// MileageGPSAtBeginningOfLoading int `db:"mileage_gps_at_beginning_of_loading" json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	// MileageGPSEmpty                int `db:"mileage_gps_empty" json:"mileage_gps_empty"`                                     // пробег в порожнем состоянии
}

func (r RowSessionObjData) GetShiftId() int {
	return r.ShiftId
}
func (r RowSessionObjData) GetSessionId() int {
	return r.SessionId
}
func (r RowSessionObjData) GetDriverId() int {
	return r.DriverId
}
func (r RowSessionObjData) GetOffset() int64 {
	return r.Offset
}
func (r RowSessionObjData) GetTimeStartSession() time.Time {
	return r.TimeStartSession
}
func (r RowSessionObjData) GetTimeUpdateSession() time.Time {
	return r.TimeUpdateSession
}
func (r RowSessionObjData) GetAvSpeed() float32 {
	return r.AvSpeed
}
func (r RowSessionObjData) GetEngHoursData() interface{} {
	return r.engHours
}
func (r RowSessionObjData) GetMileageData() interface{} {
	return r.mileageData
}
func (r RowSessionObjData) GetMileageGPSData() interface{} {
	return r.mileageGPSData
}

type responseShiftDB struct {
	data RowShiftObjData
	err  error
}

type responseSessionDB struct {
	data RowSessionObjData
	err  error
}

type responceShiftSession struct {
	responseShift   responseShiftDB
	responseSession responseSessionDB
}

// данные смены полученные от модуля
type shiftDataFromModule struct {
	mainData       dataShiftFromStorage
	engHoursData   engHoursDataInterface
	mileageData    mileageDataInterface
	mileageGPSData mileageDataInterface
}

func (s *shiftDataFromModule) loadData(data interface{}) error {
	var err error
	s.mainData, err = utils.TypeConversion[dataShiftFromStorage](data)
	if err != nil {
		err := utils.Wrapper(fmt.Errorf("ошибка конвертации основных данных"), err)
		return err
	}
	s.engHoursData, err = utils.TypeConversion[engHoursDataInterface](s.mainData.GetEngHoursData())
	if err != nil {
		err := utils.Wrapper(fmt.Errorf("ошибка данных по моточасам"), err)
		return err
	}
	s.mileageData, err = utils.TypeConversion[mileageDataInterface](s.mainData.GetMileageData())
	if err != nil {
		err := utils.Wrapper(fmt.Errorf("ошибка данных по пробегу"), err)
		return err
	}
	s.mileageGPSData, err = utils.TypeConversion[mileageDataInterface](s.mainData.GetMileageGPSData())
	if err != nil {
		err := utils.Wrapper(fmt.Errorf("ошибка данных по пробегу GPS"), err)
		return err
	}
	return err
}

// данные сессии полученные от модуля
type sessionDataFromModule struct {
	mainData       dataDriverSessionFromStorage
	engHoursData   engHoursDataInterface
	mileageData    mileageDataInterface
	mileageGPSData mileageDataInterface
}

func (s *sessionDataFromModule) loadData(data interface{}) error {
	var err error
	s.mainData, err = utils.TypeConversion[dataDriverSessionFromStorage](data)
	if err != nil {
		return err
	}
	s.engHoursData, err = utils.TypeConversion[engHoursDataInterface](s.mainData.GetEngHoursData())
	if err != nil {
		return err
	}
	s.mileageData, err = utils.TypeConversion[mileageDataInterface](s.mainData.GetMileageData())
	if err != nil {
		return err
	}
	s.mileageGPSData, err = utils.TypeConversion[mileageDataInterface](s.mainData.GetMileageGPSData())
	return err
}
