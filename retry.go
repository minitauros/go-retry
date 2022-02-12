package retry

import (
	"context"
	"time"
)

// Retry retries the given callback at max the given number of times.
// It stops as soon as a `nil` error is returned.
func Retry(numTimes int, cb func() error) error {
	var err error
	for i := 0; i <= numTimes; i++ {
		err = cb()
		if err == nil {
			break
		}
	}
	return err
}

// RetryCtx retries the given callback at max the given number of times.
// It stops as soon as a `nil` error is returned.
func RetryCtx(ctx context.Context, numTimes int, cb func() error) error {
	return Retry(numTimes, func() error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return cb()
	})
}

// RetryWithDelay retries the given callback at max the given number of times.
// It stops as soon as a `nil` error is returned.
// It sleeps for the given delay if an error happens.
func RetryWithDelay(numTimes int, delay time.Duration, cb func() error) error {
	var err error
	for i := 0; i <= numTimes; i++ {
		err = cb()
		if err == nil {
			break
		}
		time.Sleep(delay)
	}
	return err
}

// RetryWithDelayCtx retries the given callback at max the given number of times.
// It stops as soon as a `nil` error is returned.
// It sleeps for the given delay if an error happens.
func RetryWithDelayCtx(ctx context.Context, numTimes int, delay time.Duration, cb func() error) error {
	return RetryWithDelay(numTimes, delay, func() error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return cb()
	})
}

// RetryWithStop retries the given callback at max the given number of times.
// It stops only when `stop` is called.
func RetryWithStop(numTimes int, cb func(stop func()) error) error {
	var err error
	var cancelled bool
	stop := func() {
		cancelled = true
	}
	for i := 0; i <= numTimes; i++ {
		if cancelled {
			break
		}
		err = cb(stop)
	}
	return err
}

// RetryWithStopCtx retries the given callback at max the given number of times.
// It stops only when `stop` is called.
func RetryWithStopCtx(ctx context.Context, numTimes int, cb func(stop func()) error) error {
	return RetryWithStop(numTimes, func(stop func()) error {
		if ctx.Err() != nil {
			stop()
			return ctx.Err()
		}
		return cb(stop)
	})
}
