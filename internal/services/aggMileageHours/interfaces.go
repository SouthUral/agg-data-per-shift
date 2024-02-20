package aggmileagehours

type incomingMessage interface {
	GetOffset() int
	GetMessage() []byte
}
