package redisprom_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/globocom/go-redis-prometheus"
)

func TestOptions(t *testing.T) {
	assert := assert.New(t)

	t.Run("return default options", func(t *testing.T) {
		assert.Equal(&redisprom.Options{
			Namespace:       "",
			DurationBuckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}, redisprom.DefaultOptions())
	})

	t.Run("merge default options with custom values", func(t *testing.T) {
		// arrange
		custom1 := func(options *redisprom.Options) { options.Namespace = "custom" }
		custom2 := func(options *redisprom.Options) { options.DurationBuckets = []float64{0.1} }

		options := redisprom.DefaultOptions()

		// act
		options.Merge(custom1, custom2)

		// assert
		assert.Equal(&redisprom.Options{
			Namespace:       "custom",
			DurationBuckets: []float64{0.1},
		}, options)
	})

	t.Run("customize metrics namespace", func(t *testing.T) {
		// arrange
		options := redisprom.DefaultOptions()

		// act
		redisprom.WithNamespace("custom")(options)

		// assert
		assert.Equal("custom", options.Namespace)
	})

	t.Run("customize metrics duration buckets", func(t *testing.T) {
		// arrange
		options := redisprom.DefaultOptions()

		// act
		redisprom.WithDurationBuckets([]float64{0.01})(options)

		// assert
		assert.Equal([]float64{0.01}, options.DurationBuckets)
	})
}
