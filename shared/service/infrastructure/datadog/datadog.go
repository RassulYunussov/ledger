package datadog

import (
	"time"

	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Datadog interface {
	Increment(name string, tags ...string)
	Count(name string, value int64, tags ...string)
	Timing(name string, duration time.Duration, tags ...string)
	Gauge(name string, value float64, tags ...string)
	shutdown()
}

type datadog struct {
	log     *zap.Logger
	app     string
	env     string
	version string
}

func (d *datadog) extendTags(tags []string) []string {
	return append(tags, "version:"+d.version, "env:"+d.env)
}

func DatadogTracerStart(environment string, applicationName string, version string) {
	tracer.Start(
		tracer.WithRuntimeMetrics(),
		tracer.WithEnv(environment),
		tracer.WithService(applicationName),
		tracer.WithServiceVersion(version),
	)
}

func DatadogTracerStop() {
	tracer.Stop()
}
