package shiftloader

import (
	"agg-data-per-shift/pkg/structs"
	"context"
)

type storage interface {
	GetShiftsData(context.Context) (structs.ShiftSettingsData, error)
}

// type responseStorage interface {
// 	GetOffsetShifts() time.Time
// 	GetShiftsSettings() []structs.ShiftTimeSettings
// }
