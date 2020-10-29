# go-redis-prometheus

<p>
  <img src="https://img.shields.io/github/workflow/status/globocom/go-redis-prometheus/Go?style=flat-square">
  <a href="https://github.com/globocom/go-redis-prometheus/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/globocom/go-buffer?color=blue&style=flat-square">
  </a>
  <img src="https://img.shields.io/github/go-mod/go-version/globocom/go-redis-prometheus?style=flat-square">
  <a href="https://pkg.go.dev/github.com/globocom/go-redis-prometheus">
    <img src="https://img.shields.io/badge/Go-reference-blue?style=flat-square">
  </a>
</p>

[go-redis](https://github.com/go-redis/redis) hook that exports Prometheus metrics.

## Installation

    go get github.com/globocom/go-redis-prometheus

## Usage

```golang
package main

import (                                                         
    "github.com/go-redis/redis/v8"
    "github.com/globocom/go-redis-prometheus"
)

func main() {
    hook := redisprom.NewHook(
        redisprom.WithNamespace("my_namespace"),
        redisprom.WithDurationBuckets([]float64{.001, .005, .01}),
    )

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
    })
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
