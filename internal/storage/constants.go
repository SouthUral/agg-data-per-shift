package storage

const (
	// отправители сообщений
	aggMileageHours  = "aggMileageHours"
	amqp             = "amqp"
	getOffsetForAmqp = "GetOffset"

	// типы сообщений от модуля aggMileageHours

	restoreShiftDataPerObj      = "restoreShiftDataPerObj"
	addNewShiftAndSession       = "addNewShiftAndSession"
	updateShiftAndAddNewSession = "updateShiftAndAddNewSession"
	updateShiftAndSession       = "updateShiftAndSession"

	// типы ошибок

	// ошибка для модуля, который отправил сообщение в модуль storage
	commonError = "commonError"
	// критическая ошибка, модуль storage завершает работу
	criticalError = "criticalError"
)
