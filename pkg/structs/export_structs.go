package structs

import (
	"time"
)

type ShiftTimeSettings struct {
	NumShift      int
	StartShift    time.Duration
	DurationShift time.Duration
}

type ShiftSettingsData struct {
	OffsetTimeShift time.Duration
	ShiftsData      []ShiftTimeSettings
}
