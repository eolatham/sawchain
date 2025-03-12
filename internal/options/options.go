package options

import (
	"errors"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/utilities"
)

// TODO: test

const (
	errNil      = "options is nil"
	errRequired = "required argument(s) not provided"
)

// Options is a common struct for options used in Sawchain operations.
type Options struct {
	Timeout  time.Duration   // Timeout for eventual assertions.
	Interval time.Duration   // Polling interval for eventual assertions.
	Template string          // Template file path or content for Chainsaw resource operations.
	Bindings map[string]any  // Template bindings for Chainsaw resource operations.
	Object   client.Object   // Object to store state for single-resource operations.
	Objects  []client.Object // Slice to store state for multi-resource operations.
}

// parse parses variable arguments into an Options struct.
//   - If includeDurations is true, checks for Timeout and Interval; otherwise disallows them.
//   - If includeObject is true, checks for Object; otherwise disallows it.
//   - If includeObjects is true, checks for Objects; otherwise disallows it.
func parse(
	includeDurations, includeObject, includeObjects bool,
	args ...interface{},
) (*Options, error) {
	opts := &Options{}

	for _, arg := range args {
		if includeDurations {
			// Check for Timeout and Interval
			if d, ok := utilities.AsDuration(arg); ok {
				if opts.Timeout != 0 && opts.Interval != 0 {
					return nil, fmt.Errorf("too many duration arguments provided")
				} else if opts.Timeout == 0 {
					opts.Timeout = d
				} else {
					opts.Interval = d
				}
				continue
			}
		}

		if includeObject {
			// Check for Object
			if obj, ok := utilities.AsObject(arg); ok {
				if opts.Object != nil {
					return nil, fmt.Errorf("multiple client.Object arguments provided")
				} else {
					opts.Object = obj
				}
				continue
			}
		}

		if includeObjects {
			// Check for Objects
			if objs, ok := utilities.AsSliceOfObjects(arg); ok {
				if opts.Objects != nil {
					return nil, fmt.Errorf("multiple []client.Object arguments provided")
				} else {
					opts.Objects = objs
				}
				continue
			}
		}

		// Check for Template
		if str, ok := arg.(string); ok {
			if opts.Template != "" {
				return nil, fmt.Errorf("multiple template arguments provided")
			} else if utilities.IsExistingFile(str) {
				content, err := utilities.ReadFileContent(str)
				if err != nil {
					return nil, fmt.Errorf("failed to read template file: %v", err)
				}
				opts.Template = content
			} else {
				opts.Template = str
			}
			continue
		}

		// Check for Bindings
		if bindings, ok := utilities.AsMapStringAny(arg); ok {
			opts.Bindings = utilities.MergeMaps(opts.Bindings, bindings)
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

// requireTemplateOrObject requires options Template or Object to be provided.
func requireTemplateOrObject(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if opts.Template == "" && opts.Object == nil {
		return errors.New(errRequired + ": Template (string) or Object (client.Object)")
	}
	return nil
}

// requireTemplateOrObjects requires options Template or Objects to be provided.
func requireTemplateOrObjects(opts *Options) error {
	if opts == nil {
		return errors.New(errNil)
	}
	if opts.Template == "" && opts.Objects == nil {
		return errors.New(errRequired + ": Template (string) or Objects ([]client.Object)")
	}
	return nil
}

// applyDefaults applies defaults to the given options where needed.
func applyDefaults(defaults, opts *Options) *Options {
	// Default durations
	if opts.Timeout == 0 {
		opts.Timeout = defaults.Timeout
	}
	if opts.Interval == 0 {
		opts.Interval = defaults.Interval
	}

	// Merge bindings
	opts.Bindings = utilities.MergeMaps(defaults.Bindings, opts.Bindings)

	return opts
}

// parseAndApplyDefaults parses variable arguments into an Options struct
// and applies defaults where needed.
func parseAndApplyDefaults(
	defaults *Options,
	includeDurations, includeObject, includeObjects bool,
	args ...interface{},
) (*Options, error) {
	opts, err := parse(includeDurations, includeObject, includeObjects, args...)
	if err != nil {
		return nil, err
	}
	return applyDefaults(defaults, opts), nil
}

// ParseAndRequireGlobal parses and requires options
// for the Sawchain constructor.
func ParseAndRequireGlobal(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, true, false, false, args...)
	if err != nil {
		return nil, err
	}
	if err := requireDurations(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireImmediateSingle parses and requires options
// for Sawchain immediate single-resource operations.
func ParseAndRequireImmediateSingle(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, false, true, false, args...)
	if err != nil {
		return nil, err
	}
	if err := requireTemplateOrObject(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireImmediateMultiple parses and requires options
// for Sawchain immediate multi-resource operations.
func ParseAndRequireImmediateMultiple(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, false, false, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireTemplateOrObjects(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireEventualSingle parses and requires options
// for Sawchain eventual single-resource operations.
func ParseAndRequireEventualSingle(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, true, true, false, args...)
	if err != nil {
		return nil, err
	}
	if err := requireDurations(opts); err != nil {
		return nil, err
	}
	if err := requireTemplateOrObject(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// ParseAndRequireEventualMultiple parses and requires options
// for Sawchain eventual multi-resource operations.
func ParseAndRequireEventualMultiple(defaults *Options, args ...interface{}) (*Options, error) {
	opts, err := parseAndApplyDefaults(defaults, true, false, true, args...)
	if err != nil {
		return nil, err
	}
	if err := requireDurations(opts); err != nil {
		return nil, err
	}
	if err := requireTemplateOrObjects(opts); err != nil {
		return nil, err
	}
	return opts, nil
}
