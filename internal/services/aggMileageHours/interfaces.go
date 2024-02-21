package aggmileagehours

type incomingMessage interface {
	GetOffset() int64
	GetMsg() []byte
}
