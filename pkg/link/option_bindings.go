package link

// Bindings is a map of Chainsaw template bindings.
type Bindings map[string]any

// WithBinding adds to the Link's map of bindings to be used with Chainsaw templates.
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

func (b Bindings) ApplyToLink(opts LinkOptions) LinkOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToGet(opts GetOptions) GetOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}

func (b Bindings) ApplyToCheck(opts CheckOptions) CheckOptions {
	opts.Bindings = b.MergeInto(opts.Bindings)
	return opts
}
