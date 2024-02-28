package psql

// интерфейс входящего сообщения от других модулей
type trunsportMes interface {
	GetSender() string      // имя модуля отправителя сообщения
	GetMesage() interface{} // сообщение от модуля
	SendAnswer(interface{}) // метод для отправки ответа
}
