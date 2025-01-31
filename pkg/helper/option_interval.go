package helper

import "time"

// Interval is a polling interval to be used in Eventually and Consistently calls.
type Interval time.Duration

// WithInterval tells the helper how often it should poll for conditions to be met.
func WithInterval(t time.Duration) Interval {
	return Interval(t)
}

func (i Interval) ApplyToHelper(opts HelperOptions) HelperOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToCreate(opts CreateOptions) CreateOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToUpdate(opts UpdateOptions) UpdateOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToDelete(opts DeleteOptions) DeleteOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts.Interval = i
	return opts
}

func (i Interval) ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions {
	opts.Interval = i
	return opts
}
