package link

import "time"

// Interval is a polling interval to be used in Eventually and Consistently calls.
type Interval time.Duration

// WithInterval tells the Link how often it should poll for conditions to be met.
// It accepts either a time.Duration or a string (e.g. "10s", "2m").
func WithInterval(t interface{}) Interval {
	switch v := t.(type) {
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(err)
		}
		return Interval(d)
	case time.Duration:
		return Interval(v)
	default:
		panic("interval must be a string or time.Duration")
	}
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
