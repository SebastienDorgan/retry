package retry

import (
	"math"
	"sync/atomic"
	"time"
)

//Action defines an action to retry against
type Action func() (interface{}, error)

//Condition defines the retry loop stopping condition
type Condition func(interface{}, error) bool

//BackoffStrategy returns the delay to apply between attempt and attempt + 1
type BackoffStrategy func(attempt uint64, interval time.Duration) time.Duration

//Retry defines necessary attributes to retry an action
type Retry struct {
	Action          Action
	Condition       Condition
	Duration        time.Duration
	Interval        time.Duration
	MaxAtt          uint64
	BackoffStrategy BackoffStrategy
	timeout         bool
	lastValue       interface{}
	lastError       error
	attempt         uint64
}

//Result stores the result of the retry loop
type Result struct {
	Timeout   bool
	LastValue interface{}
	LastError error
	Attempts  uint64
}

//With set the action to retry
func With(action Action) *Retry {
	return &Retry{
		Action:          action,
		Condition:       FalseCondition,
		attempt:         0,
		MaxAtt:          math.MaxUint64,
		BackoffStrategy: UniformStrategy,
		Duration:        time.Second * math.MaxInt32,
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

//WithBackoff set the backoff strategy to implement
func (r *Retry) WithBackoff(strategy BackoffStrategy) *Retry {
	r.BackoffStrategy = strategy
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
		Timeout:   r.timeout,
		LastValue: r.lastValue,
		LastError: r.lastError,
		Attempts:  r.attempt,
	}
}

func (r *Retry) actionWrapper(stop *atomic.Value) {
	if stop.Load().(bool) {
		return
	}
	r.lastValue, r.lastError = r.Action()
	if r.Condition(r.lastValue, r.lastError) {
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
		d := r.BackoffStrategy(r.attempt, r.Interval)
		r.attempt++
		time.Sleep(d)
	}

}

//NoError NoError wraps a function that do not return error
func NoError(action func() interface{}) Action {
	return func() (interface{}, error) {
		return action(), nil
	}
}
