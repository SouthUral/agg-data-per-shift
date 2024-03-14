package utils

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

// структура для измерения времени процессов
type ProcessingTimeMeter struct {
	incomingCh    chan trunsportToProcessingTime
	cancel        func()
	storageResult map[string]*dataForTimeAnalitics
}

func (p *ProcessingTimeMeter) GetAnaliticsForAllProcess() {
	for process, data := range p.storageResult {
		log.Infof("Аналитика по %s: \n%+v\n ", process, data)
	}
}

func InitProcessingTimeMeter() *ProcessingTimeMeter {
	ctx, cancel := context.WithCancel(context.Background())
	p := &ProcessingTimeMeter{
		incomingCh:    make(chan trunsportToProcessingTime, 20),
		cancel:        cancel,
		storageResult: make(map[string]*dataForTimeAnalitics, 20),
	}
	go p.process(ctx)
	return p
}

// метод закрывает горутину структуры ProcessingTimeMeter
func (p *ProcessingTimeMeter) Shudown() {
	p.cancel()
}

func (p *ProcessingTimeMeter) SendMessToTimeMeter(msg trunsportToProcessingTime) {
	p.incomingCh <- msg
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

func (p *ProcessingTimeMeter) timeProcessing(msg trunsportToProcessingTime) {
	analiticsData, ok := p.storageResult[msg.nameProcess]
	if !ok {
		p.storageResult[msg.nameProcess] = initDataForTimeAnalitics(msg.GetTimeInt())
		return
	}

	analiticsData.AddValue(msg.GetTimeInt())
}

type trunsportToProcessingTime struct {
	nameProcess string
	timeProcess time.Duration
}

func (t *trunsportToProcessingTime) GetTimeInt() int64 {
	return t.timeProcess.Microseconds()
}

type dataForTimeAnalitics struct {
	minTime     int64
	maxTime     int64
	sumTime     int64
	averageTime int64
	counterData int
}

func initDataForTimeAnalitics(t int64) *dataForTimeAnalitics {
	return &dataForTimeAnalitics{
		minTime:     t,
		maxTime:     t,
		averageTime: t,
		sumTime:     t,
		counterData: 1,
	}
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
