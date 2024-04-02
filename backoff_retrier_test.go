package retry

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// I am not a mathematician and had quite a lack of sleep when I wrote this.
func expectedTimeBackedOff(initialSleepDur time.Duration, backoffCoefficient float64, numSleeps int) time.Duration {
	sleepDur := initialSleepDur
	totalSleepDur := sleepDur
	// num sleeps - 1 because the first sleep is not a multiplication, but just an initializer.
	for i := 0; i < numSleeps-1; i++ {
		sleepDur = time.Duration(math.Round(float64(sleepDur) * backoffCoefficient))
		totalSleepDur += sleepDur
	}
	return sleepDur
}

func Test_BackoffRetrier_Retry(t *testing.T) {
	Convey("*BackoffRetrier.Retry()", t, func() {
		retrier := &BackOffRetrier{
			initialDelay:       time.Millisecond,
			backOffCoefficient: 2,
		}
		var numCalled int

		Convey("If nil is returned right away, does not retry", func() {
			err := retrier.Retry(10, func() error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 1)
		})

		Convey("Retries errors until nil is returned", func() {
			startTime := time.Now()

			err := retrier.Retry(10, func() error {
				numCalled++
				if numCalled == 3 {
					return nil
				}
				return errors.New("foo")
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 3)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			startTime := time.Now()

			expectedErr := errors.New("foo")
			err := retrier.Retry(3, func() error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 4)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})
	})
}

func Test_BackoffRetrier_RetryCtx(t *testing.T) {
	Convey("*BackoffRetrier.RetryCtx()", t, func() {
		retrier := &BackOffRetrier{
			initialDelay:       time.Millisecond,
			backOffCoefficient: 2,
		}
		var numCalled int

		Convey("If nil is returned right away, does not retry", func() {
			err := retrier.RetryCtx(context.Background(), 10, func() error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 1)
		})

		Convey("Retries errors until nil is returned", func() {
			startTime := time.Now()

			err := retrier.RetryCtx(context.Background(), 10, func() error {
				numCalled++
				if numCalled == 2 {
					return nil
				}
				return errors.New("foo")
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 2)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			startTime := time.Now()

			expectedErr := errors.New("foo")
			err := retrier.RetryCtx(context.Background(), 1, func() error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})

		Convey("If the context returns an error, returns err", func() {
			startTime := time.Now()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := retrier.RetryCtx(ctx, 10, func() error {
				numCalled++
				return errors.New("foo")
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, context.Canceled)
			So(numCalled, ShouldEqual, 0)

			timeElapsed := time.Now().Sub(startTime)
			So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Error is returned immediately; no sleep.
		})
	})
}

func Test_BackoffRetrier_RetryWithStop(t *testing.T) {
	Convey("*BackoffRetrier.RetryWithStop()", t, func() {
		retrier := &BackOffRetrier{
			initialDelay:       time.Millisecond,
			backOffCoefficient: 2,
		}
		var numCalled int

		Convey("If nil is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			startTime := time.Now()

			err := retrier.RetryWithStop(2, func(stop func()) error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 3)

			timeElapsed := time.Now().Sub(startTime)
			So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Returning nil does not trigger sleep.
		})

		Convey("If err is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			startTime := time.Now()

			expectedErr := errors.New("foo")
			err := retrier.RetryWithStop(2, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 3)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})

		Convey("Stops as soon as stop is called", func() {
			Convey("If there is no error, returns nil", func() {
				startTime := time.Now()

				err := retrier.RetryWithStop(10, func(stop func()) error {
					numCalled++
					if numCalled == 2 {
						stop()
					}
					return nil
				})
				So(err, ShouldBeNil)
				So(numCalled, ShouldEqual, 2)

				timeElapsed := time.Now().Sub(startTime)
				So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Returning nil does not trigger sleep.
			})

			Convey("If there is an error, returns it", func() {
				startTime := time.Now()

				expectedErr := errors.New("foo")
				err := retrier.RetryWithStop(10, func(stop func()) error {
					numCalled++
					if numCalled == 2 {
						stop()
						return expectedErr
					}
					return nil
				})
				So(err, ShouldNotBeNil)
				So(expectedErr, ShouldEqual, expectedErr)
				So(numCalled, ShouldEqual, 2)

				timeElapsed := time.Now().Sub(startTime)
				So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Returning nil does not trigger sleep.
			})
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			startTime := time.Now()

			expectedErr := errors.New("foo")
			err := retrier.RetryWithStop(1, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})
	})
}

func Test_BackoffRetrier_RetryWithStopCtx(t *testing.T) {
	Convey("*BackoffRetrier.RetryWithStopCtx()", t, func() {
		retrier := &BackOffRetrier{
			initialDelay:       time.Millisecond,
			backOffCoefficient: 2,
		}
		var numCalled int

		Convey("If nil is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			startTime := time.Now()

			err := retrier.RetryWithStopCtx(context.Background(), 2, func(stop func()) error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 3)

			timeElapsed := time.Now().Sub(startTime)
			So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Returning nil does not trigger sleep.
		})

		Convey("If err is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			startTime := time.Now()

			expectedErr := errors.New("foo")
			err := retrier.RetryWithStopCtx(context.Background(), 2, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 3)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})

		Convey("Stops as soon as stop is called", func() {
			Convey("If there is no error, returns nil", func() {
				startTime := time.Now()

				err := retrier.RetryWithStopCtx(context.Background(), 10, func(stop func()) error {
					numCalled++
					if numCalled == 2 {
						stop()
					}
					return nil
				})
				So(err, ShouldBeNil)
				So(numCalled, ShouldEqual, 2)

				timeElapsed := time.Now().Sub(startTime)
				So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Returning nil does not trigger sleep.
			})

			Convey("If there is an error, returns it", func() {
				startTime := time.Now()

				expectedErr := errors.New("foo")
				err := retrier.RetryWithStopCtx(context.Background(), 10, func(stop func()) error {
					numCalled++
					if numCalled == 2 {
						stop()
						return expectedErr
					}
					return nil
				})
				So(err, ShouldNotBeNil)
				So(expectedErr, ShouldEqual, expectedErr)
				So(numCalled, ShouldEqual, 2)

				timeElapsed := time.Now().Sub(startTime)
				So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Returning nil does not trigger sleep.
			})
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			startTime := time.Now()

			expectedErr := errors.New("foo")
			err := retrier.RetryWithStopCtx(context.Background(), 1, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)

			timeElapsed := time.Now().Sub(startTime)
			expectedMinTimeElapsed := expectedTimeBackedOff(retrier.initialDelay, retrier.backOffCoefficient, numCalled-1) // -1 because if called 3 times then there must have been 2 sleeps.
			So(timeElapsed, ShouldBeGreaterThanOrEqualTo, expectedMinTimeElapsed)
		})

		Convey("If the context returns an error, returns err", func() {
			startTime := time.Now()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := retrier.RetryWithStopCtx(ctx, 10, func(stop func()) error {
				numCalled++
				return errors.New("foo")
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, context.Canceled)
			So(numCalled, ShouldEqual, 0)

			timeElapsed := time.Now().Sub(startTime)
			So(timeElapsed, ShouldBeLessThan, retrier.initialDelay) // Error is returned immediately; no sleep.
		})
	})
}
