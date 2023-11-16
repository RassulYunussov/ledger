package config

import (
	"flag"
	"os"
	"shared/app"
	"time"

	"go.uber.org/config"
	"go.uber.org/fx"
)

type Configuration struct {
	app.Application
	Temporal *struct {
		Namespace string
		TaskQueue string
	}
	Domain struct {
		Transfer struct {
			CurrencyCache struct {
				Capacity   uint64
				TtlSeconds time.Duration
			}
			AccountsCache struct {
				Capacity   uint64
				TtlSeconds time.Duration
			}
			CreateTransaction struct {
				TimeoutMs      time.Duration
				CircuitBreaker struct {
					MaxRequest          uint32
					ConsecutiveFailures uint32
					IntervalMs          time.Duration
					TimeoutMs           time.Duration
				}
			}
		}
	}
	InMemory *struct {
		Outpost struct {
			IntervalSeconds int
			OffsetSeconds   int
		}
		Operations struct {
			Queues int
			Buffer int
		}
		AccountCurrencies struct {
			Subsystems int
			Queues     int
			Buffer     int
			Strategy   string
		}
		Workers int
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		PoolSize int
	}
	Server struct {
		Port int
	}
	Datadog struct {
		Host    string
		Port    int
		Rate    float64
		Enabled bool
	}
}

func LoadConfiguration() (Configuration, error) {
	var c Configuration

	configurationFile := flag.String("config", "config.yaml", "configuration file")
	flag.Parse()

	cfg, err := config.NewYAML(config.File(*configurationFile), config.Expand(os.LookupEnv))
	if err != nil {
		return c, err
	}

	if err := cfg.Get("").Populate(&c); err != nil {
		return c, err
	}

	return c, nil
}

func GetApplication(c Configuration) app.Application {
	return c.Application
}

var ConfigurationModule = fx.Module("application-configuration", fx.Provide(LoadConfiguration, GetApplication))
