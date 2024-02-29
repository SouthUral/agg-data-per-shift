package storage

import (
	"time"
)

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

	EngHoursStart   float32 `db:"eng_hours_start" json:"eng_hours_start"`     // моточасы на начало
	EngHoursCurrent float32 `db:"eng_hours_current" json:"eng_hours_current"` // последняя обновленноая запись моточасов
	EngHoursEnd     float32 `db:"eng_hours_end" json:"eng_hours_end"`         // моточасы на конец

	MileageStart                int `db:"mileage_start" json:"mileage_start"`                                     // пробег на начало (смены/сессии)
	MileageCurrent              int `db:"mileage_current" json:"mileage_current"`                                 // текущий пробег
	MileageEnd                  int `db:"mileage_end" json:"mileage_end"`                                         // проебег на конец смены
	MileageLoaded               int `db:"mileage_loaded" json:"mileage_loaded"`                                   // пробег в груженом состоянии
	MileageAtBeginningOfLoading int `db:"mileage_at_beginning_of_loading" json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageEmpty                int `db:"mileage_empty" json:"mileage_empty"`                                     // пробег в порожнем состоянии

	MileageGPSStart                int `db:"mileage_gps_start" json:"mileage_gps_start"`                                     // пробег на начало (смены/сессии)
	MileageGPSCurrent              int `db:"mileage_gps_current" json:"mileage_gps_current"`                                 // текущий пробег
	MileageGPSEnd                  int `db:"mileage_gps_end" json:"mileage_gps_end"`                                         // проебег на конец смены
	MileageGPSLoaded               int `db:"mileage_gps_loaded" json:"mileage_gps_loaded"`                                   // пробег в груженом состоянии
	MileageGPSAtBeginningOfLoading int `db:"mileage_gps_at_beginning_of_loading" json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageGPSEmpty                int `db:"mileage_gps_empty" json:"mileage_gps_empty"`                                     // пробег в порожнем состоянии
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

	EngHoursStart   float32 `db:"eng_hours_start" json:"eng_hours_start"`     // моточасы на начало
	EngHoursCurrent float32 `db:"eng_hours_current" json:"eng_hours_current"` // последняя обновленноая запись моточасов
	EngHoursEnd     float32 `db:"eng_hours_end" json:"eng_hours_end"`         // моточасы на конец

	MileageStart                int `db:"mileage_start" json:"mileage_start"`                                     // пробег на начало (смены/сессии)
	MileageCurrent              int `db:"mileage_current" json:"mileage_current"`                                 // текущий пробег
	MileageEnd                  int `db:"mileage_end" json:"mileage_end"`                                         // проебег на конец смены
	MileageLoaded               int `db:"mileage_loaded" json:"mileage_loaded"`                                   // пробег в груженом состоянии
	MileageAtBeginningOfLoading int `db:"mileage_at_beginning_of_loading" json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageEmpty                int `db:"mileage_empty" json:"mileage_empty"`                                     // пробег в порожнем состоянии

	MileageGPSStart                int `db:"mileage_gps_start" json:"mileage_gps_start"`                                     // пробег на начало (смены/сессии)
	MileageGPSCurrent              int `db:"mileage_gps_current" json:"mileage_gps_current"`                                 // текущий пробег
	MileageGPSEnd                  int `db:"mileage_gps_end" json:"mileage_gps_end"`                                         // проебег на конец смены
	MileageGPSLoaded               int `db:"mileage_gps_loaded" json:"mileage_gps_loaded"`                                   // пробег в груженом состоянии
	MileageGPSAtBeginningOfLoading int `db:"mileage_gps_at_beginning_of_loading" json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageGPSEmpty                int `db:"mileage_gps_empty" json:"mileage_gps_empty"`                                     // пробег в порожнем состоянии
}

type responseDB struct {
	data []byte
	err  error
}
