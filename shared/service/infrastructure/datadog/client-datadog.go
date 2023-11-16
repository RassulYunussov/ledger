package datadog

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"go.uber.org/zap"
)

type clientDatadog struct {
	datadog
	dogstatsdClient *statsd.Client
	rate            float64
}

func (d *clientDatadog) Increment(name string, tags ...string) {
	if err := d.dogstatsdClient.Incr(d.app+"."+name, d.extendTags(tags), d.rate); err != nil {
		d.log.Error("error emitting incr metric", zap.Error(err))
	}
}

func (d *clientDatadog) Count(name string, value int64, tags ...string) {
	if err := d.dogstatsdClient.Count(d.app+"."+name, value, d.extendTags(tags), d.rate); err != nil {
		d.log.Error("error emitting count metric", zap.Error(err))
	}
}

func (d *clientDatadog) Timing(name string, duration time.Duration, tags ...string) {
	if err := d.dogstatsdClient.Timing(d.app+"."+name, duration, d.extendTags(tags), d.rate); err != nil {
		d.log.Error("error emitting timing metric", zap.Error(err))
	}
}

func (d *clientDatadog) Gauge(name string, value float64, tags ...string) {
	if err := d.dogstatsdClient.Gauge(d.app+"."+name, value, d.extendTags(tags), d.rate); err != nil {
		d.log.Error("error emitting gauge metric", zap.Error(err))
	}
}
func (d *clientDatadog) shutdown() {
	d.log.Info("closing datadog statsd client")
	if err := d.dogstatsdClient.Close(); err != nil {
		d.log.Error("error closing datadog statsd client", zap.Error(err))
	}
}

func CreateDatadogClient(log *zap.Logger, rate float64, applicationName, environment, version, address string) (Datadog, error) {
	dogstatsdClient, err := statsd.New(address)
	if err != nil {
		return nil, err
	}
	return &clientDatadog{datadog{log, applicationName, environment, version}, dogstatsdClient, rate}, nil
}

func ShutdownDatadogClient(datadog Datadog) {
	datadog.shutdown()
}
