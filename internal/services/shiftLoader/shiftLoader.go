package shiftloader

import (
	"context"
	"fmt"
	"sync"
	"time"

	"agg-data-per-shift/pkg/structs"

	log "github.com/sirupsen/logrus"
)

// ошибка определения смены (нет смены)
type defineShiftError struct {
	eventTime time.Time
}

func (e defineShiftError) Error() string {
	return fmt.Sprintf("define shift error, there is no shift : %s", e.eventTime)
}

func InitSettingsDurationShifts(intervalBetweenQueries int, storage storage) (*SettingsDurationShifts, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	res := &SettingsDurationShifts{
		mx:      sync.RWMutex{},
		shifts:  make(map[int]settingShift),
		cancel:  cancel,
		storage: storage,
	}

	go res.loadingData(ctx, intervalBetweenQueries)

	return res, ctx
}

// настройки смены
type SettingsDurationShifts struct {
	mx              sync.RWMutex
	shifts          map[int]settingShift
	offsetTimeShift time.Duration // времянное смещение смены
	cancel          func()
	storage         storage
}

type settingShift struct {
	numShift       int           // номер смены
	startTimeShift time.Duration // время старта смены
	shiftDuration  time.Duration // продолжительность смены
}

func (s *SettingsDurationShifts) loadingData(ctx context.Context, intervalBetweenQueries int) {
	timer := time.NewTicker(time.Duration(intervalBetweenQueries) * time.Minute)
	defer timer.Stop()
	for {
		err := s.getAndProcessSHiftData(ctx)
		if err != nil {
			s.Shutdown(err)
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			continue
		}
	}
}

func (s *SettingsDurationShifts) getAndProcessSHiftData(ctx context.Context) error {
	res, err := s.storage.GetShiftsData(ctx)
	if err != nil {
		return err
	}
	s.uploadingSettings(res)
	return nil
}

func (s *SettingsDurationShifts) uploadingSettings(data structs.ShiftSettingsData) {
	s.offsetTimeShift = data.OffsetTimeShift
	for _, k := range data.ShiftsData {
		s.addShiftSetting(k.NumShift, k.DurationShift, k.StartShift)
	}
}

// метод добаления смены
// TODO: нужно сделать проверку что смена не пересекается с другими сменами
func (s *SettingsDurationShifts) addShiftSetting(numShift int, shiftDuration, startTimeShift time.Duration) {
	s.mx.Lock()
	s.shifts[numShift] = settingShift{
		numShift:       numShift,
		startTimeShift: startTimeShift,
		shiftDuration:  shiftDuration,
	}
	s.mx.Unlock()
}

// определяет границы смены на переданную дату
func (s *SettingsDurationShifts) definingShiftsForTheDay(shiftSettings settingShift, pTime time.Time) (time.Time, time.Time, error) {
	var (
		startShift, endShift time.Time
		err                  error
	)

	t, err := time.Parse(time.DateOnly, pTime.Format(time.DateOnly))
	if err != nil {
		return startShift, endShift, err
	}

	startShift = t.Add(s.offsetTimeShift)
	endShift = startShift.Add(shiftSettings.shiftDuration)
	startShift = startShift.Add(1 * time.Nanosecond)
	return startShift, endShift, err
}

// метод для определения номера и даты смены
func (s *SettingsDurationShifts) defineShift(dateEvent time.Time) (int, time.Time, error) {
	// определять смену нужно по текущей дате в событии
	var numShift int
	var dateShift time.Time
	var err error

	defineDataEvent := dateEvent

	s.mx.RLock()
	for i := 0; i < 2; i++ {
		for numShift, shiftSettings := range s.shifts {
			startShift, endShift, err := s.definingShiftsForTheDay(shiftSettings, defineDataEvent)
			if err != nil {
				return numShift, dateShift, err
			}
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

func (s *SettingsDurationShifts) Shutdown(err error) {
	log.Errorf("settings Duration Shifts has terminated its active processes due to: %s", err)
	s.cancel()
}
