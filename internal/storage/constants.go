package storage

const (
	// отправители сообщений
	aggMileageHours = "aggMileageHours"

	// типы сообщений от модуля aggMileageHours

	restoreShiftDataPerObj      = "restoreShiftDataPerObj"
	addNewShiftAndSession       = "addNewShiftAndSession"
	updateShiftAndAddNewSession = "updateShiftAndAddNewSession"
	updateShiftAndSession       = "updateShiftAndSession"

	// типы ошибок

	// ошибка для модуля, который отправил сообщение в модуль storage
	modulError = "modulError"
	// критическая ошибка, модуль storage завершает работу
	criticalError = "criticalError"
	// ошибка не критическая, либо пока не требуются, неизвестны действия которые нужно сделать при возникновении ошибки,
	// ошибка никуда не отправляется, программа просто выводит ошибку в лог и не производит далее никаких действий.
	regularError = "regularError"
)
