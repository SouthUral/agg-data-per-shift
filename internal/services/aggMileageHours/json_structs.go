package aggmileagehours

import (
	"time"
)

// TODO: попробовать сделать материнскую структуру от которой будут наследоваться остальные
// TODO: но сначала запустить сервис
type rowObjData struct {
}

// данные смены (для получения из БД)
type RowShiftObjData struct {
	Id              int       `json:"id"`                // id текущей смены
	NumShift        int       `json:"num_shift"`         // номер смены (1/2....n)
	ShiftDateStart  time.Time `json:"shift_date_start"`  // время и дата начала смены (время первого события смены)
	ShiftDateEnd    time.Time `json:"shift_date_end"`    // время окончания смены (время последнего обновления)
	ShiftDate       time.Time `json:"shift_date"`        // текущая дата смены (смена может начинаться в другую дату)
	UpdatedTime     time.Time `json:"updated_time"`      // время обновления смены
	Offset          int64     `json:"event_offset"`      // offset события, которое последнее обновило состояние смены
	CurrentDriverId int       `json:"current_driver_id"` // id текущего водителя техники (может меняться в пределах смены)
	Loaded          bool      `json:"loaded"`            // флаг, техника на момент обновления записи находилась в груженом состоянии

	EngHoursStart   float32 `json:"eng_hours_start"`   // моточасы на начало
	EngHoursCurrent float32 `json:"eng_hours_current"` // последняя обновленноая запись моточасов
	EngHoursEnd     float32 `json:"eng_hours_end"`     // моточасы на конец

	MileageStart                int `json:"mileage_start"`                   // пробег на начало (смены/сессии)
	MileageCurrent              int `json:"mileage_current"`                 // текущий пробег
	MileageEnd                  int `json:"mileage_end"`                     // проебег на конец смены
	MileageLoaded               int `json:"mileage_loaded"`                  // пробег в груженом состоянии
	MileageAtBeginningOfLoading int `json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageEmpty                int `json:"mileage_empty"`                   // пробег в порожнем состоянии

	MileageGPSStart                int `json:"mileage_gps_start"`                   // пробег на начало (смены/сессии)
	MileageGPSCurrent              int `json:"mileage_gps_current"`                 // текущий пробег
	MileageGPSEnd                  int `json:"mileage_gps_end"`                     // проебег на конец смены
	MileageGPSLoaded               int `json:"mileage_gps_loaded"`                  // пробег в груженом состоянии
	MileageGPSAtBeginningOfLoading int `json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageGPSEmpty                int `json:"mileage_gps_empty"`                   // пробег в порожнем состоянии
}

func (r RowShiftObjData) GetEngHoursData() {

}

// данные сессии (для получения из БД)
type RowSessionObjData struct {
	ShiftId           int       `json:"shift_id"`            // id смены, в которой находится сессия
	SessionId         int       `json:"id"`                  // id сессии, берется из БД
	DriverId          int       `json:"driver_id"`           // id водителя
	Offset            int64     `json:"event_offset"`        // последний записанный offset
	TimeStartSession  time.Time `json:"time_start_session"`  // время старта сессии
	TimeUpdateSession time.Time `json:"time_update_session"` // время обновления записи сессии
	AvSpeed           float32   `json:"av_speed"`            // средняя скорость водителя

	EngHoursStart   float32 `json:"eng_hours_start"`   // моточасы на начало
	EngHoursCurrent float32 `json:"eng_hours_current"` // последняя обновленноая запись моточасов
	EngHoursEnd     float32 `json:"eng_hours_end"`     // моточасы на конец

	MileageStart                int `json:"mileage_start"`                   // пробег на начало (смены/сессии)
	MileageCurrent              int `json:"mileage_current"`                 // текущий пробег
	MileageEnd                  int `json:"mileage_end"`                     // проебег на конец смены
	MileageLoaded               int `json:"mileage_loaded"`                  // пробег в груженом состоянии
	MileageAtBeginningOfLoading int `json:"mileage_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageEmpty                int `json:"mileage_empty"`                   // пробег в порожнем состоянии

	MileageGPSStart                int `json:"mileage_gps_start"`                   // пробег на начало (смены/сессии)
	MileageGPSCurrent              int `json:"mileage_gps_current"`                 // текущий пробег
	MileageGPSEnd                  int `json:"mileage_gps_end"`                     // проебег на конец смены
	MileageGPSLoaded               int `json:"mileage_gps_loaded"`                  // пробег в груженом состоянии
	MileageGPSAtBeginningOfLoading int `json:"mileage_gps_at_beginning_of_loading"` // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	MileageGPSEmpty                int `json:"mileage_gps_empty"`                   // пробег в порожнем состоянии
}
