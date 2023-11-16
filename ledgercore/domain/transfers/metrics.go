package transfers

import (
	"fmt"
	"shared/service/infrastructure/datadog"

	"github.com/google/uuid"
)

func withMetricPrefix(name string) string {
	return "transaction." + name
}

func increment(dd datadog.Datadog, name string, subsystem uuid.UUID, tags ...string) {
	tags = append(tags, fmt.Sprintf("subsystem:%v", subsystem))
	dd.Increment(withMetricPrefix(name), tags...)
}
