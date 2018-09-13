# retry

**retry** is a simple golang retrying module aiming at simplifying the creation of generic retry behavior

# Features
* Fluent interface
* Specify stop condition
* Specify timeout during
* Specify retry interavl

```golang
package retry_test

import (
	"testing"
	"time"

	"github.com/SebastienDorgan/retry"
	"github.com/stretchr/testify/assert"
)

func hello() interface{} {
	println("hello")
	time.Sleep(500 * time.Millisecond)
	return nil
}

func Counter(start, step int) retry.Action {
	value := start
	return func() interface{} {
		value = value + step
		return value
	}
}

func GreaterThan(v int) retry.Condition {
	return func(vi interface{}) bool {
		return vi.(int) >= v
	}
}

func Test(t *testing.T) {
	//run with go test -v -timeout 30s github.com/SebastienDorgan/retry -run ^Test$

	//Retry hello function every seconds for 10 seconds
	start := time.Now()

	res := retry.With(hello).Every(1 * time.Second).For(10 * time.Second).Go()

	elapse := time.Now().Sub(start)

	//The retry mechanism is under millis precise
	assert.Equal(t, 10*time.Second, elapse.Truncate(time.Millisecond))
	assert.True(t, res.Timeout)

	//Retry the counter function every 10 seconds for 10 seconds or until condition GreaterThen(10) is satisfied
	res = retry.With(Counter(0, 2)).For(10 * time.Second).Every(1 * time.Second).Until(GreaterThan(10)).Go()
	assert.Equal(t, 10, res.LastValue.(int))
	assert.False(t, res.Timeout)
}

```
