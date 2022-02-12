package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRetry(t *testing.T) {
	Convey("Retry()", t, func() {
		var numCalled int

		Convey("If nil is returned right away, does not retry", func() {
			err := Retry(10, func() error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 1)
		})

		Convey("Retries errors until nil is returned", func() {
			err := Retry(10, func() error {
				numCalled++
				if numCalled == 2 {
					return nil
				}
				return errors.New("foo")
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 2)
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			expectedErr := errors.New("foo")
			err := Retry(1, func() error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)
		})
	})
}

func TestRetryCtx(t *testing.T) {
	Convey("RetryCtx()", t, func() {
		var numCalled int

		Convey("If nil is returned right away, does not retry", func() {
			err := RetryCtx(context.Background(), 10, func() error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 1)
		})

		Convey("Retries errors until nil is returned", func() {
			err := RetryCtx(context.Background(), 10, func() error {
				numCalled++
				if numCalled == 2 {
					return nil
				}
				return errors.New("foo")
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 2)
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			expectedErr := errors.New("foo")
			err := RetryCtx(context.Background(), 1, func() error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)
		})

		Convey("If the context returns an error, returns err", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := RetryCtx(ctx, 10, func() error {
				numCalled++
				return errors.New("foo")
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, context.Canceled)
			So(numCalled, ShouldEqual, 0)
		})
	})
}

func TestRetryWithDelay(t *testing.T) {
	Convey("RetryWithDelay()()", t, func() {
		var numCalled int

		Convey("If nil is returned right away, does not retry", func() {
			err := RetryWithDelay(10, 0, func() error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 1)
		})

		Convey("Retries errors until nil is returned", func() {
			startTime := time.Now()
			delay := 10 * time.Millisecond
			err := RetryWithDelay(10, delay, func() error {
				numCalled++
				if numCalled == 3 {
					return nil
				}
				return errors.New("foo")
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 3)
			So(time.Since(startTime), ShouldBeGreaterThan, 2*delay)
			So(time.Since(startTime), ShouldBeLessThan, 3*delay)
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			expectedErr := errors.New("foo")
			startTime := time.Now()
			delay := 10 * time.Millisecond
			err := RetryWithDelay(1, delay, func() error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)
			So(time.Since(startTime), ShouldBeGreaterThan, 2*delay)
			So(time.Since(startTime), ShouldBeLessThan, 3*delay)
		})
	})
}

func TestRetryWithDelayCtx(t *testing.T) {
	Convey("RetryCtx()", t, func() {
		var numCalled int

		Convey("If nil is returned right away, does not retry", func() {
			err := RetryWithDelayCtx(context.Background(), 10, 0, func() error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 1)
		})

		Convey("Retries errors until nil is returned", func() {
			startTime := time.Now()
			delay := 10 * time.Millisecond
			err := RetryWithDelayCtx(context.Background(), 10, delay, func() error {
				numCalled++
				if numCalled == 3 {
					return nil
				}
				return errors.New("foo")
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 3)
			So(time.Since(startTime), ShouldBeGreaterThan, 2*delay)
			So(time.Since(startTime), ShouldBeLessThan, 3*delay)
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			startTime := time.Now()
			delay := 10 * time.Millisecond
			expectedErr := errors.New("foo")
			err := RetryWithDelayCtx(context.Background(), 1, delay, func() error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)
			So(time.Since(startTime), ShouldBeGreaterThan, 2*delay)
			So(time.Since(startTime), ShouldBeLessThan, 3*delay)
		})

		Convey("If the context returns an error, returns err", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := RetryWithDelayCtx(ctx, 10, 0, func() error {
				numCalled++
				return errors.New("foo")
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, context.Canceled)
			So(numCalled, ShouldEqual, 0)
		})
	})
}

func TestRetryWithStop(t *testing.T) {
	Convey("RetryWithStop()", t, func() {
		var numCalled int

		Convey("If nil is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			err := RetryWithStop(3, func(stop func()) error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 4)
		})

		Convey("If err is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			expectedErr := errors.New("foo")
			err := RetryWithStop(3, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 4)
		})

		Convey("Stops as soon as stop is called", func() {
			Convey("If there is no error, returns nil", func() {
				err := RetryWithStop(10, func(stop func()) error {
					numCalled++
					if numCalled == 2 {
						stop()
					}
					return nil
				})
				So(err, ShouldBeNil)
				So(numCalled, ShouldEqual, 2)
			})

			Convey("If there is an error, returns it", func() {
				expectedErr := errors.New("foo")
				err := RetryWithStop(10, func(stop func()) error {
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
			})
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			expectedErr := errors.New("foo")
			err := RetryWithStop(1, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)
		})
	})
}

func TestRetryWithStopCtx(t *testing.T) {
	Convey("RetryWithStopCtx()", t, func() {
		var numCalled int

		Convey("If nil is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			err := RetryWithStopCtx(context.Background(), 3, func(stop func()) error {
				numCalled++
				return nil
			})
			So(err, ShouldBeNil)
			So(numCalled, ShouldEqual, 4)
		})

		Convey("If err is returned, but stop is not called, keeps retrying until the maximum number of retries is reached", func() {
			expectedErr := errors.New("foo")
			err := RetryWithStopCtx(context.Background(), 3, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 4)
		})

		Convey("Stops as soon as stop is called", func() {
			Convey("If there is no error, returns nil", func() {
				err := RetryWithStopCtx(context.Background(), 10, func(stop func()) error {
					numCalled++
					if numCalled == 2 {
						stop()
					}
					return nil
				})
				So(err, ShouldBeNil)
				So(numCalled, ShouldEqual, 2)
			})

			Convey("If there is an error, returns it", func() {
				expectedErr := errors.New("foo")
				err := RetryWithStopCtx(context.Background(), 10, func(stop func()) error {
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
			})
		})

		Convey("If the maximum number of tries is reached, returns err", func() {
			expectedErr := errors.New("foo")
			err := RetryWithStopCtx(context.Background(), 1, func(stop func()) error {
				numCalled++
				return expectedErr
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, expectedErr)
			So(numCalled, ShouldEqual, 2)
		})

		Convey("If the context returns an error, returns err", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := RetryWithStopCtx(ctx, 10, func(stop func()) error {
				numCalled++
				return errors.New("foo")
			})
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, context.Canceled)
			So(numCalled, ShouldEqual, 0)
		})
	})
}
