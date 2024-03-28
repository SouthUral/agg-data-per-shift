package storage

import (
	"context"
	"time"

	"agg-data-per-shift/pkg/structs"
)

type shiftSettingsHandler struct {
	dbConn *PgConn
}

func (s shiftSettingsHandler) GetShiftsData(ctx context.Context) (structs.ShiftSettingsData, error) {
	var response structs.ShiftSettingsData

	rows, err := s.dbConn.QueryDB(ctx, getSettingsShifts)
	if err != nil {
		return response, err
	}

	convRows, err := convertQueryRows[ResponceShiftsData](rows)
	if err != nil {
		return response, err
	}

	response = converToResponse(convRows)
	return response, err
}

func converToResponse(rows []ResponceShiftsData) structs.ShiftSettingsData {
	res := structs.ShiftSettingsData{
		OffsetTimeShift: rows[0].StartOffset,
		ShiftsData:      make([]structs.ShiftTimeSettings, 0, len(rows)),
	}

	for _, k := range rows {
		item := structs.ShiftTimeSettings{
			NumShift:      k.Number,
			StartShift:    k.Start,
			DurationShift: k.Duration,
		}
		res.ShiftsData = append(res.ShiftsData, item)
	}
	return res
}

type ResponceShiftsData struct {
	StartOffset time.Duration `db:"start_offset"`
	Number      int           `db:"number"`
	Description string        `db:"description"`
	Start       time.Duration `db:"start"`
	Duration    time.Duration `db:"duration"`
}
