package retry

import (
	"context"
	"math"
	"time"
)

// Retrier is a function for retrying the given callback at max the given number of times.
// stop must be called to stop retrying.
type Retrier func(numTimes int, cb func(stop func()) error) error

// RetrierCtx is a function for retrying the given callback at max the given number of times.
// stop must be called to stop retrying.
type RetrierCtx func(ctx context.Context, numTimes int, cb func(stop func()) error) error

// BackOffRetrier retries a given callback, backing off on failure.
type BackOffRetrier struct {
	initialDelay       time.Duration
	backOffCoefficient float64
}

// NewBackOffRetrier returns a new back off retrier.
func NewBackOffRetrier(initialDelay time.Duration, backOffCoefficient float64) *BackOffRetrier {
	return &BackOffRetrier{initialDelay: initialDelay, backOffCoefficient: backOffCoefficient}
}

// Retry retries the given callback at max the given number of times.
// It stops as soon as a `nil` error is returned.
func (r *BackOffRetrier) Retry(numTimes int, cb func() error) error {
	delay := r.initialDelay
	return Retry(numTimes, func() error {
		err := cb()
		if err != nil {
			time.Sleep(delay)
			delay = time.Duration(math.Round(r.backOffCoefficient * float64(delay)))
		}
		return err
	})
}

// RetryCtx retries the given callback at max the given number of times.
// It stops as soon as a `nil` error is returned.
func (r *BackOffRetrier) RetryCtx(ctx context.Context, numTimes int, cb func() error) error {
	delay := r.initialDelay
	return Retry(numTimes, func() error {
		err := ctx.Err()
		if err != nil {
			return err
		}

		err = cb()
		if err != nil {
			time.Sleep(delay)
			delay = time.Duration(math.Round(r.backOffCoefficient * float64(delay)))
		}
		return err
	})
}

// RetryWithStop retries the given callback at max the given number of times.
// It stops only when `stop` is called.
func (r *BackOffRetrier) RetryWithStop(numTimes int, cb func(stop func()) error) error {
	delay := r.initialDelay
	return RetryWithStop(numTimes, func(stop func()) error {
		err := cb(stop)
		if err != nil {
			time.Sleep(delay)
			delay = time.Duration(math.Round(r.backOffCoefficient * float64(delay)))
		}
		return err
	})
}

// RetryWithStopCtx retries the given callback at max the given number of times.
// It stops only when `stop` is called.
func (r *BackOffRetrier) RetryWithStopCtx(ctx context.Context, numTimes int, cb func(stop func()) error) error {
	delay := r.initialDelay
	return RetryWithStop(numTimes, func(stop func()) error {
		err := ctx.Err()
		if err != nil {
			stop()
			return err
		}

		err = cb(stop)
		if err != nil {
			time.Sleep(delay)
			delay = time.Duration(math.Round(r.backOffCoefficient * float64(delay)))
		}
		return err
	})
}
