package cmd

import (
	"github.com/spf13/viper"
)

type Config struct {
	Verbose bool `mapstructure:"verbose"`
	Schema string `mapstructure:"schema"`
	Out string `mapstructure:"out"`
	Exclude []string `mapstructure:"exclude"`
	Include []string `mapstructure:"include"`
	Src string `mapstructure:"src"`
	Template string `mapstructure:"template"`
}

func initConfigFile(path string, args *Args) error {
	if path == "" {
		return nil
	}

	vip := viper.New()

	// # Read os env
	vip.AutomaticEnv()

	// # Tell viper the path/location of your config file
	vip.SetConfigFile(path)

	// # Viper reads all the variables from env file and log error if any found
	if err := vip.ReadInConfig(); err != nil {
		return err
	}

	config := &Config{}

	// # Viper unmarshals the loaded env varialbes into the struct
	if err := vip.Unmarshal(config); err != nil {
		return err
	}

	args.Verbose = config.Verbose
	args.LoaderParams.Schema = config.Schema
	args.OutParams.Out = config.Out
	if len(config.Exclude) > 0 {
		for _, exclude := range config.Exclude {
			if err := args.SchemaParams.Exclude.Set(exclude); err != nil {
				return err
			}
		}
	}
	if len(config.Include) > 0 {
		for _, include := range config.Include {
			if err := args.SchemaParams.Include.Set(include); err != nil {
				return err
			}
		}
	}
	args.TemplateParams.Src = config.Src

	return nil
}