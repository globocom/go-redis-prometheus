package redisprom

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	// Hook represents a go-redis hook that exports metrics of commands and pipelines.
	//
	// The following metrics are exported:
	//
	// - Single commands (not-pipelined)
	//   - Histogram of duration
	//   - Counter of errors
	//
	// - Pipelined commands
	//   - Counter of commands
	//   - Counter of errors
	//
	// The duration of individual pipelined commands won't be collected, but the overall duration of the
	// pipeline will, with a pseudo-command called "pipeline".
	Hook struct {
		options           *Options
		singleCommands    *prometheus.HistogramVec
		pipelinedCommands *prometheus.CounterVec
		singleErrors      *prometheus.CounterVec
		pipelinedErrors   *prometheus.CounterVec
	}

	startKey struct{}
)

var (
	labelNames = []string{"instance", "command"}
)

// NewHook creates a new go-redis hook instance and registers Prometheus collectors.
func NewHook(opts ...Option) *Hook {
	options := DefaultOptions()
	options.Merge(opts...)

	singleCommands := register(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Name:      "redis_single_commands",
		Help:      "Histogram of single Redis commands",
		Buckets:   options.DurationBuckets,
	}, labelNames)).(*prometheus.HistogramVec)

	pipelinedCommands := register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Name:      "redis_pipelined_commands",
		Help:      "Number of pipelined Redis commands",
	}, labelNames)).(*prometheus.CounterVec)

	singleErrors := register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Name:      "redis_single_errors",
		Help:      "Number of single Redis commands that have failed",
	}, labelNames)).(*prometheus.CounterVec)

	pipelinedErrors := register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Name:      "redis_pipelined_errors",
		Help:      "Number of pipelined Redis commands that have failed",
	}, labelNames)).(*prometheus.CounterVec)

	return &Hook{
		options:           options,
		singleCommands:    singleCommands,
		pipelinedCommands: pipelinedCommands,
		singleErrors:      singleErrors,
		pipelinedErrors:   pipelinedErrors,
	}
}

func (hook *Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		duration := time.Since(start).Seconds()
		hook.singleCommands.WithLabelValues(hook.options.InstanceName, cmd.Name()).Observe(duration)
	}

	if isActualErr(cmd.Err()) {
		hook.singleErrors.WithLabelValues(hook.options.InstanceName, cmd.Name()).Inc()
	}

	return nil
}

func (hook *Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if err := hook.AfterProcess(ctx, redis.NewCmd(ctx, "pipeline")); err != nil {
		return err
	}

	for _, cmd := range cmds {
		hook.pipelinedCommands.WithLabelValues(hook.options.InstanceName, cmd.Name()).Inc()

		if isActualErr(cmd.Err()) {
			hook.pipelinedErrors.WithLabelValues(hook.options.InstanceName, cmd.Name()).Inc()
		}
	}

	return nil
}

func register(collector prometheus.Collector) prometheus.Collector {
	err := prometheus.DefaultRegisterer.Register(collector)
	if err == nil {
		return collector
	}

	if arErr, ok := err.(prometheus.AlreadyRegisteredError); ok {
		return arErr.ExistingCollector
	}

	panic(err)
}

func isActualErr(err error) bool {
	return err != nil && err != redis.Nil
}
