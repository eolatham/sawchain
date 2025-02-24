package link

// TemplateFile is a path to a Chainsaw template file.
type TemplateFile string

// WithTemplateFile tells the Link to read from a Chainsaw template file
// at the given path. State will still be written to the given struct.
func WithTemplateFile(path string) TemplateFile {
	return TemplateFile(path)
}

func (t TemplateFile) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts.TemplateFile = t
	return opts
}

func (t TemplateFile) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts.TemplateFile = t
	return opts
}

func (t TemplateFile) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts.TemplateFile = t
	return opts
}

func (t TemplateFile) ApplyToGet(opts GetOptions) GetOptions {
	opts.TemplateFile = t
	return opts
}

func (t TemplateFile) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts.TemplateFile = t
	return opts
}

func (t TemplateFile) ApplyToCheck(opts CheckOptions) CheckOptions {
	opts.TemplateFile = t
	return opts
}
