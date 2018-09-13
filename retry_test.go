package retry_test

import (
	"testing"
	"time"

	"github.com/CS-SI/retry"
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
	res := retry.With(hello).For(10 * time.Second).Every(1 * time.Second).Go()
	println(res.Timeout)
	res = retry.With(Counter(0, 2)).For(10 * time.Second).Every(1 * time.Second).Until(GreaterThan(10)).Go()
	println(res.LastValue.(int))
	println(res.Timeout)
}
