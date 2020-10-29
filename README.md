# go-redis-prometheus

[go-redis](https://github.com/go-redis/redis) hook to export Prometheus metrics.

## Installing

    go get github.com/globocom/go-redis-prometheus

## Usage

```golang
package main

import ()

func main() {
    hook := redisprom.NewHook(
        redisprom.WithNamespace("my_namespace"),
        redisprom.WithDurationBuckets([]float64{.001, .005, .01},
    )

    client := redis.NewClient()
    client.AddHook(hook)

    // run redis commands...
}
```

## Exported metrics

The hook exports the following metrics:

- Single commands (not pipelined):
  - Histogram of commands: `redis_single_commands{name="command"}`
  - Counter of errors: `redis_single_errors{name="command"}`
 - Pipelined commands:
   - Counter of commands: `redis_pipelined_commands{name="command"}`
   - Counter of errors: `redis_pipelined_errors{name="command"}`

## Note on pipelines

It isn't possible to measure the duration of individual
pipelined commands, but the duration of the pipeline itself is calculated and 
exported as a pseudo-command called "pipeline" under the single command metric.
