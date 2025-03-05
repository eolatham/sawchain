package options

import (
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/utilities"
)

// TODO: test

type Options struct {
	Timeout  time.Duration
	Interval time.Duration
	Template string
	Bindings map[string]any
	Object   client.Object
	Objects  []client.Object
}

// Parse parses variable arguments into an Options struct.
// If includeDurations is true, checks for Timeout and Interval; otherwise disallows them.
// If includeObject is true, checks for Object; otherwise disallows it.
// If includeObjects is true, checks for Objects; otherwise disallows it.
func Parse(includeDurations, includeObject, includeObjects bool, args ...interface{}) (*Options, error) {
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
			if obj, ok := utilities.AsClientObject(arg); ok {
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
			if objs, ok := utilities.AsSliceOfClientObjects(arg); ok {
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

// ApplyDefaults applies defaults to the given options where needed.
func ApplyDefaults(defaults, opts *Options) *Options {
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

// ParseAndApplyDefaults parses variable arguments into an Options struct and applies defaults where needed.
func ParseAndApplyDefaults(
	includeDurations, includeObject, includeObjects bool, defaults *Options, args ...interface{},
) (*Options, error) {
	opts, err := Parse(includeDurations, includeObject, includeObjects, args...)
	if err != nil {
		return nil, err
	}
	return ApplyDefaults(defaults, opts), nil
}
