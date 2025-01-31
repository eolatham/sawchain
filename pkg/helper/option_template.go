package helper

// Template is a path to a Chainsaw template file.
type Template string

// WithTemplate tells the helper to read from a Chainsaw template file
// at the given path. State will still be written to the given struct.
func WithTemplate(template string) Template {
	return Template(template)
}

func (t Template) ApplyToCreate(opts CreateOptions) CreateOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToUpdate(opts UpdateOptions) UpdateOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToDelete(opts DeleteOptions) DeleteOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions {
	opts.Template = t
	return opts
}
