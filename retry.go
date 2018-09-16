package retry

import (
	"math"
	"sync/atomic"
	"time"
)

//Action defines an action to retry against
type Action func() interface{}

//Condition defines the retry loop stopping condition
type Condition func(interface{}) bool

//Retry defines necessary attributes to retry an action
type Retry struct {
	Action    Action
	Condition Condition
	Duration  time.Duration
	Interval  time.Duration
	MaxAtt    uint64
	timeout   bool
	lastValue interface{}
	attempt   uint64
}

//Result stores the result of the retry loop
type Result struct {
	Timeout           bool
	LastReturnedValue interface{}
	Attempts          uint64
}

//With set the action to retry
func With(action Action) *Retry {
	return &Retry{
		Action:    action,
		Condition: FalseCondition,
		attempt:   0,
		MaxAtt:    math.MaxUint64,
	}

}

//Until set the stopping condition
func (r *Retry) Until(condition Condition) *Retry {
	r.Condition = condition
	return r
}

//Every set the retry interval
func (r *Retry) Every(duration time.Duration) *Retry {
	r.Interval = duration
	return r
}

//For set the maximum duration of the retry loop
func (r *Retry) For(duration time.Duration) *Retry {
	r.Duration = duration
	return r
}

//MaxAttempts set the maximum attempts of the retry loop
func (r *Retry) MaxAttempts(max uint64) *Retry {
	r.MaxAtt = max
	return r
}

//Go starts the retry loop
func (r *Retry) Go() *Result {
	end := make(chan bool, 1)
	var stop atomic.Value
	stop.Store(false)
	go r.loop(end, &stop)
	select {
	case <-end:
		break
	case <-time.After(r.Duration):
		stop.Store(true)
		r.timeout = true
	}
	return &Result{
		Timeout:           r.timeout,
		LastReturnedValue: r.lastValue,
		Attempts:          r.attempt,
	}
}

func (r *Retry) actionWrapper(stop *atomic.Value) {
	if stop.Load().(bool) {
		return
	}
	r.lastValue = r.Action()
	if r.Condition(r.lastValue) {
		stop.Store(true)
	}
}

func (r *Retry) loop(end chan bool, stop *atomic.Value) {
	for {
		if r.attempt >= r.MaxAtt {
			end <- true
			break
		}
		if stop.Load().(bool) {
			end <- true
			break
		}
		go r.actionWrapper(stop)
		r.attempt++

		time.Sleep(r.Interval)
	}

}
