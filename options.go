package redisprom

type (
	// Options represents options to customize the exported metrics.
	Options struct {
		Namespace       string
		DurationBuckets []float64
	}

	Option func(*Options)
)

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		Namespace:       "",
		DurationBuckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	}
}

func (options *Options) Merge(opts ...Option) {
	for _, opt := range opts {
		opt(options)
	}
}

// WithNamespace sets the namespace of all metrics.
func WithNamespace(namespace string) Option {
	return func(options *Options) {
		options.Namespace = namespace
	}
}

// WithDurationBuckets sets the duration buckets of single commands metrics.
func WithDurationBuckets(buckets []float64) Option {
	return func(options *Options) {
		options.DurationBuckets = buckets
	}
}
