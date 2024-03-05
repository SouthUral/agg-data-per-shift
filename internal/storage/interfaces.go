package storage

import (
	"time"
)

// интерфейс входящего сообщения от других модулей
type trunsportMes interface {
	GetSender() string                  // имя модуля отправителя сообщения
	GetMesage() interface{}             // сообщение от модуля
	GetChForResponse() chan interface{} // метод для отправки ответа
}

// интерфейс сообщения от модуля aggMileageHours
type mesFromAggMileageHours interface {
	GetType() string
	GetObjID() int
	GetShiftData() interface{}
	GetSessionData() interface{}
}

// данные смены
type dataShiftFromStorage interface {
	GetShiftId() int  // возвращает Id смены (индентификатор в таблице БД)
	GetShiftNum() int // возвращает номер смены
	GetShiftDateStart() time.Time
	GetShiftDateEnd() time.Time
	GetShiftDate() time.Time   // возвращает дату смены
	GetUpdatedTime() time.Time // время последнего обновления (пока под вопросом)
	GetOffset() int64          // offset последнего события, которое было применено к смене
	GetCurrentDriverId() int
	GetStatusLoaded() bool // груженый или нет (показатель всегда на послденее обновление записи)

	GetEngHoursData() interface{}   // получение интерфейса к данным о моточасах за смену
	GetMileageData() interface{}    // получение интерфейса к данным пробега за смену
	GetMileageGPSData() interface{} // получение интерфейса к данным пробега по GPS за смену
}

// данные текущей сессии водителя
type dataDriverSessionFromStorage interface {
	GetShiftId() int                 // id смены, в которой находится сессия
	GetSessionId() int               // id сессии
	GetDriverId() int                // id водителя
	GetOffset() int64                // offset последнего события, которое было применено к сессии
	GetTimeStartSession() time.Time  // время начала сессии
	GetTimeUpdateSession() time.Time // время последнего обновления сессии

	GetAvSpeed() float32 // средняя скорость водителя в сессии

	GetEngHoursData() interface{}   // получение интерфейса к данным о моточасах за сессию
	GetMileageData() interface{}    // получение интерфейса к данным пробега за сессию
	GetMileageGPSData() interface{} // получение интерфейса к данным пробега по GPS за сессию
}

// интерфейс получения данных о моточасах
type engHoursDataInterface interface {
	GetEngHoursStart() float32   // моточасы на начало смены/сессии
	GetEngHoursCurrent() float32 // последний обновленный показатель моточасов
	GetEngHoursEnd() float32     // моточасы на конец смены/сессии
}

// интерфейс получения данных о пробеге
type mileageDataInterface interface {
	GetMileageStart() int                // пробег на начало смены/сессии
	GetMileageCurrent() int              // пробег на последнее обновление записи
	GetMileageEnd() int                  // пробег на конец смены/сессии
	GetMileageLoaded() int               // пробег груженым
	GetMileageAtBeginningOfLoading() int // пробег на начало последней погрузки (может быть 0 если на момент чтения данных из БД машина была не грудеженая)
	GetMileageEmpty() int                // пробег порожним
}
