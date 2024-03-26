package shiftloader

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ошибка определения смены (нет смены)
type defineShiftError struct {
	eventTime time.Time
}

func (e defineShiftError) Error() string {
	return fmt.Sprintf("define shift error, there is no shift : %s", e.eventTime)
}

func initSettingsDurationShifts(offsetTimeShift int) *settingsDurationShifts {
	res := &settingsDurationShifts{
		mx:              sync.RWMutex{},
		shifts:          make(map[int]settingShift),
		offsetTimeShift: offsetTimeShift,
	}

	return res
}

// настройки смены
type settingsDurationShifts struct {
	mx              sync.RWMutex
	shifts          map[int]settingShift
	offsetTimeShift int // времянное смещение смены
}

type settingShift struct {
	numShift       int       // номер смены
	startTimeShift time.Time // время старта смены
	shiftDuration  int       // продолжительность смены
}

func (s *settingsDurationShifts) loadingData(ctx context.Context, intervalBetweenQueries int) {
	timer := time.NewTicker(time.Duration(intervalBetweenQueries) * time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			// здесь должен ьыть запрос в storage
		}
	}
}

// метод добаления смены
// TODO: нужно сделать проверку что смена не пересекается с другими сменами
func (s *settingsDurationShifts) AddShiftSetting(numShift, shiftDuration int, startTimeShift time.Time) {
	s.mx.Lock()
	s.shifts[numShift] = settingShift{
		numShift:       numShift,
		startTimeShift: startTimeShift,
		shiftDuration:  shiftDuration,
	}
	s.mx.Unlock()
}

// определяет границы смены на переданную дату
func (s *settingsDurationShifts) definingShiftsForTheDay(shiftSettings settingShift, pTime time.Time) (time.Time, time.Time) {
	t := time.Date(pTime.Year(),
		pTime.Month(),
		pTime.Day(),
		shiftSettings.startTimeShift.Hour(),
		shiftSettings.startTimeShift.Minute(),
		shiftSettings.startTimeShift.Second(),
		shiftSettings.startTimeShift.Nanosecond(),
		pTime.Location(),
	)
	startShift := t.Add(time.Duration(s.offsetTimeShift) * time.Hour)
	endShift := startShift.Add(time.Duration(shiftSettings.shiftDuration) * time.Hour)
	startShift = startShift.Add(1 * time.Nanosecond)
	return startShift, endShift
}

// метод для определения номера и даты смены
func (s *settingsDurationShifts) defineShift(dateEvent time.Time) (int, time.Time, error) {
	// определять смену нужно по текущей дате в событии
	var numShift int
	var dateShift time.Time
	var err error

	defineDataEvent := dateEvent

	s.mx.RLock()
	for i := 0; i < 2; i++ {
		for numShift, shiftSettings := range s.shifts {
			startShift, endShift := s.definingShiftsForTheDay(shiftSettings, defineDataEvent)
			compare := dateEvent.After(startShift) && dateEvent.Before(endShift)
			if compare {
				return numShift, endShift, err
			}
		}
		defineDataEvent = defineDataEvent.Add(24 * time.Hour)
	}

	defer s.mx.RUnlock()

	err = defineShiftError{dateEvent}

	return numShift, dateShift, err
}
