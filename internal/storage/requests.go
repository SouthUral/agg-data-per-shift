package storage

const (
	// запросы в БД для сообщений от модуля aggMileageHours
	// запросы на получение последней смены по objectId
	getLastObjShift = "SELECT * FROM reports.shifts WHERE object_id = $1 ORDER BY id LIMIT 1"
	// запрос на получение последней сессии по object_id
	getLastObjSession = "SELECT * FROM reports.drivers_sessions WHERE object_id = $1 ORDER BY id LIMIT 1"
)
