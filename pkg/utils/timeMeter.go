package utils

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// структура для измерения времени процессов
type ProcessingTimeMeter struct {
	incomingCh    chan TrunsportToProcessingTime
	cancel        func()
	storageResult map[string]*dataForTimeAnalitics
	mx            *sync.RWMutex
}

func (p *ProcessingTimeMeter) GetAnaliticsForAllProcess() {
	for process, data := range p.storageResult {
		log.Infof("Аналитика по %s: \n%+v\n ", process, data)
	}
}

func InitProcessingTimeMeter() (*ProcessingTimeMeter, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &ProcessingTimeMeter{
		incomingCh:    make(chan TrunsportToProcessingTime, 100),
		cancel:        cancel,
		storageResult: make(map[string]*dataForTimeAnalitics, 20),
		mx:            &sync.RWMutex{},
	}
	go p.process(ctx)
	return p, ctx
}

// метод закрывает горутину структуры ProcessingTimeMeter
func (p *ProcessingTimeMeter) Shudown() {
	p.cancel()
}

// возвращает счетчик выбранного измерителя
func (p *ProcessingTimeMeter) GetCounterOnKey(key string) (int, bool) {
	p.mx.RLock()
	data, ok := p.storageResult[key]
	if !ok {
		return 0, ok
	}
	return data.GetCounter(), ok
}

func (p *ProcessingTimeMeter) SendMessToTimeMeter(msg TrunsportToProcessingTime) {
	select {
	case p.incomingCh <- msg:
		return
	}
}

func (p *ProcessingTimeMeter) process(ctx context.Context) {
	defer log.Warning("ProcessingTimeMeter has finished")

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-p.incomingCh:
			p.timeProcessing(msg)
		}
	}

}

func (p *ProcessingTimeMeter) timeProcessing(msg TrunsportToProcessingTime) {
	analiticsData, ok := p.storageResult[msg.NameProcess]
	if !ok {
		p.storageResult[msg.NameProcess] = initDataForTimeAnalitics(msg.GetTimeInt())
		return
	}

	analiticsData.AddValue(msg.GetTimeInt())
}

type TrunsportToProcessingTime struct {
	NameProcess string
	TimeProcess time.Duration
}

func (t *TrunsportToProcessingTime) GetTimeInt() int64 {
	return t.TimeProcess.Microseconds()
}

type dataForTimeAnalitics struct {
	minTime     int64
	maxTime     int64
	sumTime     int64
	averageTime int64
	counterData int
	mx          *sync.RWMutex
}

func initDataForTimeAnalitics(t int64) *dataForTimeAnalitics {
	return &dataForTimeAnalitics{
		minTime:     t,
		maxTime:     t,
		averageTime: t,
		sumTime:     t,
		counterData: 1,
		mx:          &sync.RWMutex{},
	}
}

func (d *dataForTimeAnalitics) GetCounter() int {
	d.mx.RLock()
	res := d.counterData
	d.mx.RUnlock()
	return res
}

func (d *dataForTimeAnalitics) SetMinTime(t int64) {
	if t < d.minTime {
		d.minTime = t
	}
}

func (d *dataForTimeAnalitics) SetMaxTime(t int64) {
	if t > d.maxTime {
		d.maxTime = t
	}
}

func (d *dataForTimeAnalitics) SetAverageTime(t int64) {
	d.sumTime += t
	d.counterData += 1
	d.averageTime = d.sumTime / int64(d.counterData)
}

func (d *dataForTimeAnalitics) AddValue(t int64) {
	d.SetMinTime(t)
	d.SetMaxTime(t)
	d.SetAverageTime(t)
}
