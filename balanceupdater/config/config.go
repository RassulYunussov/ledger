package config

import (
	"flag"
	"os"
	"shared/app"

	"go.uber.org/config"
	"go.uber.org/fx"
)

type Configuration struct {
	app.Application
	Temporal struct {
		Namespace string
		TaskQueue string
		Workers   int
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
