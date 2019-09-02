package retry_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SebastienDorgan/retry"
)

var countHello = 0

func hello() interface{} {
	println("hello")
	time.Sleep(500 * time.Millisecond)
	return nil
}

func Counter(start, step int) retry.Action {
	value := start
	return func() (interface{}, error) {
		value = value + step
		return value, nil
	}
}

func GreaterThan(v int) retry.Condition {
	return func(vi interface{}, e error) bool {
		return vi.(int) >= v
	}
}

func ComplexAction(shared *atomic.Value, attempt int) retry.Action {
	return func() (i interface{}, e error) {
		cpt := shared.Load().(int)
		if cpt == attempt {
			number := 42
			return &number, nil
		}
		shared.Store(cpt + 1)
		return nil, fmt.Errorf("error")
	}
}

func Test(t *testing.T) {
	//run with go test -v -timeout 120s github.com/SebastienDorgan/retry -run ^Test$

	//Retry hello function every seconds for 10 seconds
	start := time.Now()

	res := retry.With(retry.NoError(hello)).Every(1 * time.Second).For(10 * time.Second).Go()

	elapse := time.Now().Sub(start)

	//The retry mechanism is under millis precise
	assert.Equal(t, 10*time.Second, elapse.Truncate(time.Millisecond))
	assert.True(t, res.Timeout)
	assert.Equal(t, uint64(10), res.Attempts)

	//If MaxAttempts is used the retry loop stops before the timeout
	res = retry.With(retry.NoError(hello)).Every(1 * time.Second).For(10 * time.Second).MaxAttempts(5).Go()
	assert.False(t, res.Timeout)
	assert.Equal(t, uint64(5), res.Attempts)

	//Retry the counter function every 10 seconds for 10 seconds or until condition GreaterThen(10) is satisfied
	res = retry.With(Counter(0, 2)).For(10 * time.Second).Every(1 * time.Second).Until(GreaterThan(10)).Go()
	assert.Equal(t, 10, res.LastValue.(int))
	assert.False(t, res.Timeout)

	//Retry hello function every seconds for 10 seconds
	start = time.Now()

	res = retry.With(Counter(0, 1)).Every(1 * time.Second).WithBackoff(retry.ExponentialStrategy(2.)).MaxAttempts(5).Go()

	elapse = time.Now().Sub(start)
	assert.Equal(t, 5, res.LastValue.(int))
	//31 = 1*2^0 + 1*2^1 + 1*2^2 + 1*2^3 + 1*2^4
	assert.Equal(t, 31*time.Second, elapse.Truncate(time.Second))
	v := atomic.Value{}
	cpt := 0
	v.Store(cpt)
	action := ComplexAction(&v, 5)
	condition := func(v interface{}, err error) bool {
		return v != nil
	}
	ret := retry.With(action).For(5 * time.Minute).Every(10 * time.Second).Until(condition).Go()
	assert.Equal(t, uint64(6), ret.Attempts)
	assert.NoError(t, ret.LastError)
	assert.Equal(t, 42, *ret.LastValue.(*int))
	assert.False(t, ret.Timeout)
}
