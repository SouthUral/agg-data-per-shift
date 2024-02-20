package aggmileagehours

// события для отправки в горутину агрегации
type eventForAgg struct {
	offset    int
	eventData *eventData
}
