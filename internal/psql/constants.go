package psql

const (
	// отправители сообщений
	aggMileageHours = "aggMileageHours"

	// типы сообщений от модуля aggMileageHours
	restoreShiftDataPerObj      = "restoreShiftDataPerObj"
	addNewShiftAndSession       = "addNewShiftAndSession"
	updateShiftAndAddNewSession = "updateShiftAndAddNewSession"
	updateShiftAndSession       = "updateShiftAndSession"
)
