package datadog

import (
	"time"

	"go.uber.org/zap"
)

func (d *datadog) Increment(name string, tags ...string) {
	d.log.Debug("emulating emiting increment metric", zap.String("name", name), zap.Any("with tags", d.extendTags(tags)))
}

func (d *datadog) Count(name string, value int64, tags ...string) {
	d.log.Debug("emulating emiting count metric", zap.String("name", name), zap.Int64("value", value), zap.Any("with tags", d.extendTags(tags)))
}

func (d *datadog) Timing(name string, duration time.Duration, tags ...string) {
	d.log.Debug("emulating emiting timing metric", zap.String("name", name), zap.Duration("duration", duration), zap.Any("with tags", d.extendTags(tags)))
}

func (d *datadog) Gauge(name string, value float64, tags ...string) {
	d.log.Debug("emulating emiting gauge metric", zap.String("name", name), zap.Float64("duration", value), zap.Any("with tags", d.extendTags(tags)))
}

func CreateNoopDatadogClient(log *zap.Logger, rate float64, applicationName, environment, version, address string) (Datadog, error) {
	return &datadog{log, applicationName, environment, version}, nil
}

func (d *datadog) shutdown() {
	d.log.Debug("emulationg closing datadog statsd client")
}
