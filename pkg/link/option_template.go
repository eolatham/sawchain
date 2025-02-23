package link

// Template is a path to a Chainsaw template file.
type Template string

// WithTemplate tells the Link to read from a Chainsaw template file
// at the given path. State will still be written to the given struct.
func WithTemplate(template string) Template {
	return Template(template)
}

func (t Template) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToGet(opts GetOptions) GetOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts.Template = t
	return opts
}

func (t Template) ApplyToCheck(opts CheckOptions) CheckOptions {
	opts.Template = t
	return opts
}
