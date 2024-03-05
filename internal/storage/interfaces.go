package storage

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
	GetShiftData() []byte
	GetSessionData() []byte
}
