package options

import (
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/utilities"
)

// TODO: add other variations of options

// Options struct
type Options struct {
	EventuallyTimeout         time.Duration
	EventuallyPollingInterval time.Duration
	TemplateFile              string
	TemplateContent           string
	TemplateBindings          map[string]any
	Object                    client.Object
}

// ParseOptions parses variable arguments into an Options struct.
func ParseOptions(args ...interface{}) (*Options, error) {
	opts := &Options{
		TemplateBindings: make(map[string]any),
	}

	for _, arg := range args {
		// Check for time.Duration options
		if d, ok := utilities.AsDuration(arg); ok {
			if opts.EventuallyTimeout == 0 {
				opts.EventuallyTimeout = d
			} else if opts.EventuallyPollingInterval == 0 {
				opts.EventuallyPollingInterval = d
			} else {
				return nil, fmt.Errorf("too many duration arguments provided")
			}
			continue
		}

		// Check for client.Object
		if obj, ok := utilities.AsClientObject(arg); ok {
			if opts.Object != nil {
				return nil, fmt.Errorf("multiple client.Object instances provided")
			}
			opts.Object = obj
			continue
		}

		// Check for template file
		if str, ok := arg.(string); ok {
			if opts.TemplateFile == "" && opts.TemplateContent == "" {
				opts.TemplateFile = str
			} else if opts.TemplateContent == "" {
				opts.TemplateContent = str
			} else {
				return nil, fmt.Errorf("both templateFile and templateContent provided")
			}
			continue
		}

		// Check for template bindings
		if bindings, ok := utilities.AsMapStringAny(arg); ok {
			for k, v := range bindings {
				opts.TemplateBindings[k] = v
			}
			continue
		}

		// If the argument doesn't match any expected type
		return nil, fmt.Errorf("unexpected argument type: %T", arg)
	}

	// Validate template file vs content
	if opts.TemplateFile != "" && opts.TemplateContent != "" {
		return nil, fmt.Errorf("templateFile and templateContent are mutually exclusive")
	}

	return opts, nil
}
