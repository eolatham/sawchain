package link

import "time"

// Interval is a polling interval to be used in Eventually and Consistently calls.
type Interval time.Duration

// WithInterval tells the Link how often it should poll for conditions to be met.
func WithInterval(t time.Duration) Interval {
	return Interval(t)
}

func (i Interval) ApplyToLink(opts LinkOptions) LinkOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts.Interval = i
	return opts
}
