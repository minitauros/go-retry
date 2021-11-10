The retry package provides functions for retrying any type of action a given number of times.

## Regular retry functions

### Retry()

```go
err := Retry(3, func() error {
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

### RetryCtx()

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel the context. 
// The function will never be called and 
// the context error will be returned right away.
cancel() 

err := RetryCtx(ctx, 3, func() error {
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

### RetryWithStop()

If you want to have fine grained control over when the retrying should stop.

```go
err := RetryWithStop(3, func(stop func()) error {
    err := someFunc()
    if err != nil {
        if _, ok := err.(SomeErr); ok {
            stop() // Don't retry this type of error.
            return err // Return this error.
        }

        return err // Retry.
    }
    stop() // No err was returned. We can stop retrying.
    return nil // Retry.
})
```

### RetryWithStopCtx()

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel the context. 
// The function will never be called and 
// the context error will be returned right away.
cancel()
    
err := RetryWithStopCtx(ctx, 3, func(stop func()) error {
    err := someFunc()
    if err != nil {
        if _, ok := err.(SomeErr); ok {
            stop() // Don't retry this type of error.
            return err // Return this error.
        }

        return err // Retry.
    }
    stop() // No err was returned. We can stop retrying.
    return nil // Retry.
})
```

## Retry with backoff

Works the same as the regular retry functions, but sleeps before making a new attempt.

Example with the regular `Retry()` function:

```go
retrier := NewBackOffRetrier(time.Second, 2)
err := retrier.Retry(3, func() error {
    err := someFunc()
    if err != nil {
        // Retry.
        // First attempt will sleep for time.Second.
        // Second attempt will sleep for time.Second * 2.
        // Third attempt will sleep for time.Second * 2 * 2.
        // Etc.
        // If 4th attempt (max number of attempts exceeded), 
    	// this error is returned.
        return err
    }
    return nil // Stop retrying.
})
```