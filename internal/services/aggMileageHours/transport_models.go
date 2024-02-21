package aggmileagehours

// события для отправки в горутину агрегации
type eventForAgg struct {
	offset    int64
	eventData *eventData
}
