package options

import (
	"errors"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/util"
)

const (
	errNil              = "options is nil"
	errRequired         = "required argument(s) not provided"
	errObjectAndObjects = "client.Object and []client.Object arguments both provided"
)

// Options is a common struct for options used in Sawchain operations.
type Options struct {
	Timeout  time.Duration   // Timeout for eventual assertions.
	Interval time.Duration   // Polling interval for eventual assertions.
	Template string          // Template content for Chainsaw resource operations.
	Bindings map[string]any  // Template bindings for Chainsaw resource operations.
	Object   client.Object   // Object to store state for single-resource operations.
	Objects  []client.Object // Slice to store state for multi-resource operations.
}

// parse parses variable arguments into an Options struct.
//   - If includeDurations is true, checks for Timeout and Interval; otherwise disallows them.
//   - If includeObject is true, checks for Object; otherwise disallows it.
//   - If includeObjects is true, checks for Objects; otherwise disallows it.
//   - If includeTemplate is true, checks for Template; otherwise disallows it.
func parse(
	includeDurations bool,
	includeObject bool,
	includeObjects bool,
	includeTemplate bool,
	args ...interface{},
) (*Options, error) {
	opts := &Options{
		Bindings: map[string]any{},
	}

	for _, arg := range args {
		if includeDurations {
			// Check for Timeout and Interval
			if d, ok := util.AsDuration(arg); ok {
				if opts.Timeout != 0 && opts.Interval != 0 {
					return nil, errors.New("too many duration arguments provided")
				} else if opts.Timeout == 0 {
					opts.Timeout = d
				} else if d > opts.Timeout {
					return nil, errors.New("provided interval is greater than timeout")
				} else {
					opts.Interval = d
				}
				continue
			}
		}

		if includeObject {
			// Check for Object
			if obj, ok := util.AsObject(arg); ok {
				if opts.Object != nil {
					return nil, errors.New("multiple client.Object arguments provided")
				} else if opts.Objects != nil {
					return nil, errors.New(errObjectAndObjects)
				} else if util.IsNil(obj) {
					return nil, errors.New(
						"provided client.Object is nil or has a nil underlying value")
				} else {
					opts.Object = obj
				}
				continue
			}
		}

		if includeObjects {
			// Check for Objects
			if objs, ok := util.AsSliceOfObjects(arg); ok {
				if opts.Objects != nil {
					return nil, errors.New("multiple []client.Object arguments provided")
				} else if opts.Object != nil {
					return nil, errors.New(errObjectAndObjects)
				} else if util.ContainsNil(objs) {
					return nil, errors.New(
						"provided []client.Object contains an element that is nil or has a nil underlying value")
				} else {
					opts.Objects = objs
				}
				continue
			}
		}

		if includeTemplate {
			// Check for Template
			if str, ok := arg.(string); ok {
				if opts.Template != "" {
					return nil, errors.New("multiple template arguments provided")
				} else if util.IsExistingFile(str) {
					content, err := util.ReadFileContent(str)
					if err != nil {
						return nil, fmt.Errorf("failed to read template file: %v", err)
					}
					opts.Template = content
				} else {
					opts.Template = str
				}
				continue
			}
		}

		// Check for Bindings
		if bindings, ok := util.AsMapStringAny(arg); ok {
			opts.Bindings = util.MergeMaps(opts.Bindings, bindings)
			continue
		}

		return nil, fmt.Errorf("unexpected argument type: %T", arg)
	}

	return opts, nil
}

// requireDurations requires options Timeout and Interval to be provided.
func requireDurations(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if opts.Timeout == 0 {
		return errors.New(errRequired + ": Timeout (string or time.Duration)")
	}
	if opts.Interval == 0 {
		return errors.New(errRequired + ": Interval (string or time.Duration)")
	}
	return nil
}

// requireTemplate requires option Template to be provided.
func requireTemplate(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if len(opts.Template) == 0 {
		return errors.New(errRequired + ": Template (string)")
	}
	return nil
}

// requireTemplateObject requires options Template or Object to be provided.
func requireTemplateObject(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if len(opts.Template) == 0 && opts.Object == nil {
		return errors.New(errRequired + ": Template (string) or Object (client.Object)")
	}
	return nil
}

// requireTemplateObjects requires options Template or Objects to be provided.
func requireTemplateObjects(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if len(opts.Template) == 0 && opts.Objects == nil {
		return errors.New(errRequired + ": Template (string) or Objects ([]client.Object)")
	}
	return nil
}

// requireTemplateObjectObjects requires options Template, Object, or Objects to be provided.
func requireTemplateObjectObjects(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if len(opts.Template) == 0 && opts.Object == nil && opts.Objects == nil {
		return errors.New(errRequired + ": Template (string), Object (client.Object), or Objects ([]client.Object)")
	}
	return nil
}

// applyDefaults applies defaults to the given options where needed.
func applyDefaults(defaults, opts *Options) *Options {
	// Nil checks
	if defaults == nil {
		return opts
	} else if opts == nil {
		return defaults
	}

	// Default durations
	if opts.Timeout == 0 {
		opts.Timeout = defaults.Timeout
	}
	if opts.Interval == 0 {
		opts.Interval = defaults.Interval
	}

	// Merge bindings
	opts.Bindings = util.MergeMaps(defaults.Bindings, opts.Bindings)

	return opts
}

// parseAndApplyDefaults parses variable arguments into an Options struct
// and applies defaults where needed.
func parseAndApplyDefaults(
	defaults *Options,
	includeDurations bool,
	includeObject bool,
	includeObjects bool,
	includeTemplate bool,
	args ...interface{},
) (*Options, error) {
	opts, err := parse(includeDurations, includeObject, includeObjects, includeTemplate, args...)
	if err != nil {
		return nil, err
	}
	return applyDefaults(defaults, opts), nil
}

// ParseAndRequireGlobal parses and requires options for the Sawchain constructor.
func ParseAndRequireGlobal(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, true, false, false, false, args...)
	if err != nil {
		return nil, err
	}
	if err := requireDurations(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireEventual parses and requires options for Sawchain eventual operations.
func ParseAndRequireEventual(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, true, true, true, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireDurations(opts); err != nil {
		return nil, err
	}
	if err := requireTemplateObjectObjects(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireImmediate parses and requires options for Sawchain immediate operations.
func ParseAndRequireImmediate(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, false, true, true, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireTemplateObjectObjects(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireImmediateSingle parses and requires options
// for Sawchain immediate single-resource operations.
func ParseAndRequireImmediateSingle(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, false, true, false, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireTemplateObject(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireImmediateMulti parses and requires options
// for Sawchain immediate multi-resource operations.
func ParseAndRequireImmediateMulti(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, false, false, true, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireTemplateObjects(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireImmediateTemplate parses and requires options
// for Sawchain immediate template operations.
func ParseAndRequireImmediateTemplate(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, false, true, true, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireTemplate(opts); err != nil {
		return nil, err
	}
	return opts, nil
}
