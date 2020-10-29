package redisprom

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type (
	// Hook represents a go-redis hook that exports metrics of commands and pipelines.
	//
	// The following metrics are exported:
	//
	// Single commands (not-pipelined)
	//   Histogram of duration
	//   Counter of errors
	//
	// Pipelined commands
	//   Counter of commands
	//   Counter of errors
	//
	// Duration of individual pipelined commands won't be collected, but the overall duration of the
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

var labelNames = []string{"instance", "command"}

// NewHook creates a new go-redis hook instance and registers Prometheus collectors.
func NewHook(opts ...Option) *Hook {
	options := DefaultOptions()
	options.Merge(opts...)

	singleCommands := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Name:      "redis_single_commands",
		Help:      "Histogram of single Redis commands",
		Buckets:   options.DurationBuckets,
	}, labelNames)

	pipelinedCommands := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Name:      "redis_pipelined_commands",
		Help:      "Number of pipelined Redis commands",
	}, labelNames)

	singleErrors := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Name:      "redis_single_errors",
		Help:      "Number of single Redis commands that have failed",
	}, labelNames)

	pipelinedErrors := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Name:      "redis_pipelined_errors",
		Help:      "Number of pipelined Redis commands that have failed",
	}, labelNames)

	return &Hook{options, singleCommands, pipelinedCommands, singleErrors, pipelinedErrors}
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
		if isActualErr(cmd.Err()) {
			hook.pipelinedErrors.WithLabelValues(hook.options.InstanceName, cmd.Name()).Inc()
		}
	}

	return nil
}

// Close unregisters all collectors initialized by NewHook.
func (hook *Hook) Close() error {
	prometheus.Unregister(hook.singleCommands)
	prometheus.Unregister(hook.pipelinedCommands)
	prometheus.Unregister(hook.singleErrors)
	prometheus.Unregister(hook.pipelinedErrors)
	return nil
}

func isActualErr(err error) bool {
	return err != nil && err != redis.Nil
}
