package link

import "time"

// Timeout is a polling timeout to be used in Eventually calls.
type Timeout time.Duration

// WithTimeout tells the Link how long it should wait for conditions to be met.
func WithTimeout(t time.Duration) Timeout {
	return Timeout(t)
}

func (t Timeout) ApplyToLink(opts LinkOptions) LinkOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts.Timeout = t
	return opts
}
