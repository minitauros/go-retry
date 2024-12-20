# Retry

The retry package provides functions for retrying any type of action a given number of times.

## Usage

```
go get github.com/minitauros/go-retry
```

For the rest, see the examples below.

* [Regular retry functions](#regular-retry-functions)
  * [Retry()](#retry)
  * [RetryCtx()](#retryctx)
  * [RetryWithDelay()](#retrywithdelay)
  * [RetryWithDelayCtx()](#retrywithdelayctx)
  * [RetryWithStop()](#retrywithstop)
  * [RetryWithStopCtx()](#retrywithstopctx)
* [Retry with backoff](#retry-with-backoff)

## Regular retry functions

### Retry()

```go
err := Retry(3, func() error {
    err := someFunc()
    if err != nil {
        time.Sleep(time.Second) // Wait a bit before retrying.
        // Retry.
        // If 4th attempt (max number of attempts exceeded),
        // this error is returned.
        return err
    }
    return nil // Stop retrying.
})
```

### RetryCtx()

```go
ctx, cancel := context.WithCancel(context.Background())

// If the context has an error, the callback function will
// never be called and the context error will be
// returned right away.
cancel() 

err := RetryCtx(ctx, 3, func() error {
    err := someFunc()
    if err != nil {
        time.Sleep(time.Second) // Wait a bit before retrying.
        // Retry.
        // If 4th attempt (max number of attempts exceeded),
        // this error is returned.
        return err
    }
    return nil // Stop retrying.
})
```
### RetryWithDelay()

```go
err := RetryWithDelay(3, time.Second, func() error {
    err := someFunc()
    if err != nil {
        // We will sleep for 1 second and then retry.
        // If 4th attempt (max number of attempts exceeded),
        // this error is returned.
        return err
    }
    return nil // Stop retrying.
})
```

### RetryWithDelayCtx()

```go
ctx, cancel := context.WithCancel(context.Background())

// If the context has an error, the callback function will
// never be called and the context error will be
// returned right away.
cancel() 

err := RetryWithDelayCtx(ctx, 3, time.Second, func() error {
    err := someFunc()
    if err != nil {
        // We will sleep for 1 second and then retry.
        // If 4th attempt (max number of attempts exceeded),
        // this error is returned.
        return err
    }
    return nil // Stop retrying.
})
```

### RetryWithStop()

If you want to have fine grained control over when the retrying should stop (e.g. if a certain type of error is encountered).

Returning `nil` or `err` does not stop retrying. It only stops when `stop()` is called.

If `stop()` is called and then an `err` is returned, that is the error that will be returned from the `RetryWithStop` function.

```go
err := RetryWithStop(3, func(stop func()) error {
    err := someFunc()
    if err != nil {
        if _, ok := err.(SomeErr); ok {
            stop() // Don't retry this type of error.
            return err // Return this error.
        }

        time.Sleep(time.Second) // Wait a bit before retrying.
        return err // Retry.
        return nil // This would have the same effect (also retry).
    }
    stop() // No err was returned. We can stop retrying.
    return nil
})
```

### RetryWithStopCtx()

```go
ctx, cancel := context.WithCancel(context.Background())

// If the context has an error, the callback function will
// never be called and the context error will be
// returned right away.
cancel()
    
err := RetryWithStopCtx(ctx, 3, func(stop func()) error {
    err := someFunc()
    if err != nil {
        if _, ok := err.(SomeErr); ok {
            stop() // Don't retry this type of error.
            return err // Return this error.
        }

        time.Sleep(time.Second) // Wait a bit before retrying.
        return err // Retry.
        return nil // This would have the same effect (also retry).
    }
    stop() // No err was returned. We can stop retrying.
    return nil
})
```

## Retry with backoff

Works the same as the regular retry functions, but sleeps according to specified backoff before making a new attempt.

Example with the regular `Retry()` function:

```go
// First attempt will sleep for time.Second.
// Second attempt will sleep for time.Second * 2.
// Third attempt will sleep for time.Second * 2 * 2.
// Etc.
retrier := NewBackOffRetrier(time.Second, 2)
err := retrier.Retry(3, func() error {
    err := someFunc()
    if err != nil {
        // Retry.
        // If 4th attempt (max number of attempts exceeded),
        // this error is returned.
        return err
    }
    return nil // Stop retrying.
})
```