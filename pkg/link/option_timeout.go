package link

import "time"

// Timeout is a polling timeout to be used in Eventually calls.
type Timeout time.Duration

// WithTimeout tells the Link how long it should wait for conditions to be met.
// Accepts either a time.Duration or a string (e.g. "10s", "2m").
func WithTimeout(t interface{}) Timeout {
	switch v := t.(type) {
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(err)
		}
		return Timeout(d)
	case time.Duration:
		return Timeout(v)
	default:
		panic("timeout must be a string or time.Duration")
	}
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
