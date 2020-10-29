package redisprom_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"

	"github.com/globocom/go-redis-prometheus"
)

func TestHook(t *testing.T) {
	assert := assert.New(t)
	namespace := "my_namespace"

	t.Run("create a new hook", func(t *testing.T) {
		// act
		sut := redisprom.NewHook()
		defer sut.Close()

		// assert
		assert.NotNil(sut)
	})

	t.Run("export metrics after a command is processed", func(t *testing.T) {
		// arrange
		sut := redisprom.NewHook(
			redisprom.WithNamespace(namespace),
			redisprom.WithDurationBuckets([]float64{0.1, 0.2}),
		)
		defer sut.Close()

		cmd := redis.NewStringCmd(context.TODO(), "get")
		cmd.SetErr(errors.New("some error"))

		// act
		ctx, err1 := sut.BeforeProcess(context.TODO(), cmd)
		err2 := sut.AfterProcess(ctx, cmd)

		// assert
		assert.Nil(err1)
		assert.Nil(err2)

		metrics, err := prometheus.DefaultGatherer.Gather()
		assert.Nil(err)

		assert.ElementsMatch([]string{
			"my_namespace_redis_single_commands",
			"my_namespace_redis_single_errors",
		}, filter(metrics, namespace))
	})

	t.Run("export metrics after a pipeline is processed", func(t *testing.T) {
		// arrange
		sut := redisprom.NewHook(
			redisprom.WithNamespace(namespace),
			redisprom.WithDurationBuckets([]float64{0.1, 0.2}),
		)
		defer sut.Close()

		cmd := redis.NewStringCmd(context.TODO(), "get")
		cmd.SetErr(errors.New("some error"))

		// act
		ctx, err1 := sut.BeforeProcessPipeline(context.TODO(), []redis.Cmder{cmd})
		err2 := sut.AfterProcessPipeline(ctx, []redis.Cmder{cmd})

		// assert
		assert.Nil(err1)
		assert.Nil(err2)

		metrics, err := prometheus.DefaultGatherer.Gather()
		assert.Nil(err)

		assert.ElementsMatch([]string{
			"my_namespace_redis_single_commands",
			"my_namespace_redis_pipelined_errors",
		}, filter(metrics, namespace))
	})
}

func filter(metrics []*dto.MetricFamily, namespace string) []string {
	var result []string
	for _, metric := range metrics {
		if strings.HasPrefix(*metric.Name, namespace) {
			result = append(result, *metric.Name)
		}
	}
	return result
}
