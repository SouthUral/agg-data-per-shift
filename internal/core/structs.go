package core

type msgEvent interface {
	GetTypeMsg() string
	GetReverceCh() chan interface{}
	GetMsg() []byte
	GetOffset() int64
}

// TODO: нужно переделать все под универсальную структуру
type trunsportMes interface {
	GetSender() string                  // имя модуля отправителя сообщения
	GetMesage() interface{}             // сообщение от модуля
	GetChForResponse() chan interface{} // метод для отправки ответа
}

// транспортная структура (универсальный интерфейс)
type transportStruct struct {
	sender         string           // модуль отправитель сообщения
	mesage         interface{}      // сообщение отправителя
	reverseChannel chan interface{} // канал для отправки ответа от модуля storage
}

func (t transportStruct) GetSender() string {
	return t.sender
}

func (t transportStruct) GetMesage() interface{} {
	return t.mesage
}

// метод для отправки ответа от модуля storage
func (t transportStruct) GetChForResponse() chan interface{} {
	return t.reverseChannel
}

// ответное сообщение
type answerEvent struct {
	offset int
}

func (a answerEvent) GetOffset() int {
	return a.offset
}
