package storage

// интерфейс входящего сообщения от других модулей
type trunsportMes interface {
	GetSender() string      // имя модуля отправителя сообщения
	GetMesage() interface{} // сообщение от модуля
	SendAnswer(interface{}) // метод для отправки ответа
}

// интерфейс сообщения от модуля aggMileageHours
type mesFromAggMileageHours interface {
	GetType() string
	GetObjID() int
	GetShiftData() interface{}
	GetSessionData() interface{}
}
