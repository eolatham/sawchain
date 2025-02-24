package link

// TemplateContent is a Chainsaw template string.
type TemplateContent string

// WithTemplateContent tells the Link to read from the given Chainsaw template.
// State will still be written to the given struct.
func WithTemplateContent(content string) TemplateContent {
	return TemplateContent(content)
}

func (t TemplateContent) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts.TemplateContent = t
	return opts
}

func (t TemplateContent) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts.TemplateContent = t
	return opts
}

func (t TemplateContent) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts.TemplateContent = t
	return opts
}

func (t TemplateContent) ApplyToGet(opts GetOptions) GetOptions {
	opts.TemplateContent = t
	return opts
}

func (t TemplateContent) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts.TemplateContent = t
	return opts
}

func (t TemplateContent) ApplyToCheck(opts CheckOptions) CheckOptions {
	opts.TemplateContent = t
	return opts
}
