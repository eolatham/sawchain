package helper

import "time"

// Timeout is a polling timeout to be used in Eventually and Consistently calls.
type Timeout time.Duration

// WithTimeout tells the helper how long it should wait for conditions to be met.
func WithTimeout(t time.Duration) Timeout {
	return Timeout(t)
}

func (t Timeout) ApplyToHelper(opts HelperOptions) HelperOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToCreate(opts CreateOptions) CreateOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToUpdate(opts UpdateOptions) UpdateOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToDelete(opts DeleteOptions) DeleteOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts.Timeout = t
	return opts
}

func (t Timeout) ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions {
	opts.Timeout = t
	return opts
}
