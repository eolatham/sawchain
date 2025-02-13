package helper

// Bindings is a map of Chainsaw template bindings.
type Bindings map[string]any

// WithBinding adds to the helper's map of bindings to be used with Chainsaw templates.
func WithBinding(name string, value any) Bindings {
	return map[string]any{name: value}
}

func (b Bindings) MergeInto(bindings Bindings) Bindings {
	for name, value := range b {
		if bindings == nil {
			bindings = map[string]any{name: value}
		} else {
			bindings[name] = value
		}
	}
	return bindings
}

func (b Bindings) ApplyToHelper(opts HelperOptions) HelperOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToCreate(opts CreateOptions) CreateOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToUpdate(opts UpdateOptions) UpdateOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToDelete(opts DeleteOptions) DeleteOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}
