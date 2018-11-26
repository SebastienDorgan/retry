package retry_test

import (
	"testing"
	"time"

	"github.com/SebastienDorgan/retry"
	"github.com/stretchr/testify/assert"
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

func Test(t *testing.T) {
	//run with go test -v -timeout 30s github.com/SebastienDorgan/retry -run ^Test$

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
}
