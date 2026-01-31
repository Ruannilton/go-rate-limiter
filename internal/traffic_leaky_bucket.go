package internal

import (
	"errors"
	"time"
)

type LeakyBucketTrafficShaper struct {
	ticker      *time.Ticker
	queue       chan chan bool
	closeSignal <-chan struct{}
}

func NewLeakyBucketTrafficShaper(capacity int, dropPerSecond int, closeSignal <-chan struct{}) *LeakyBucketTrafficShaper {
	interval := time.Second / time.Duration(dropPerSecond)

	shaper := &LeakyBucketTrafficShaper{
		ticker:      time.NewTicker(interval),
		queue:       make(chan chan bool, capacity),
		closeSignal: closeSignal,
	}
	go func(shaper *LeakyBucketTrafficShaper) {
		for {
			select {
			case <-shaper.ticker.C:
				select {
				case responseChannel := <-shaper.queue:
					responseChannel <- true
					close(responseChannel)
				default:
				}
			case <-shaper.closeSignal:
				shaper.ticker.Stop()
				return
			}
		}
	}(shaper)
	return shaper
}

func (l *LeakyBucketTrafficShaper) AddRequest() <-chan bool {
	responseChan := make(chan bool, 1)
	l.queue <- responseChan
	return responseChan
}

type LeakyBucketTrafficShaperParams struct {
	Capacity     int
	DropPerSecond int
}

func GetLeakyBucketTrafficShaperParamsFromMap(params map[string]any) (LeakyBucketTrafficShaperParams,error) {
	capacity, capacityOk := getNumberFromMap[int](params, "capacity")
	if !capacityOk {
		return LeakyBucketTrafficShaperParams{},errors.New("invalid capacity parameter")
	}
	dropPerSecond, dropOk := getNumberFromMap[int](params, "drop_per_second")
	if !dropOk {
		return LeakyBucketTrafficShaperParams{},errors.New("invalid drop_per_second parameter")
	}
	return LeakyBucketTrafficShaperParams{
		Capacity:     capacity,
		DropPerSecond: dropPerSecond,
	},nil
}



