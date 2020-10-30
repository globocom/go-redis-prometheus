package redisprom

type (
	// Options represents options to customize the exported metrics.
	Options struct {
		InstanceName    string
		Namespace       string
		DurationBuckets []float64
	}

	Option func(*Options)
)

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		InstanceName:    "unnamed",
		Namespace:       "",
		DurationBuckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	}
}

func (options *Options) Merge(opts ...Option) {
	for _, opt := range opts {
		opt(options)
	}
}

// WithInstanceName sets the name of the Redis instance.
func WithInstanceName(name string) Option {
	return func(options *Options) {
		options.InstanceName = name
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
