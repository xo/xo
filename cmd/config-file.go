package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Verbose bool `mapstructure:"verbose"`
	Schema string `mapstructure:"schema"`
	Out string `mapstructure:"out"`
	Exclude []string `mapstructure:"exclude"`
	Include []string `mapstructure:"include"`
	Src string `mapstructure:"src"`
}

func initConfigFile(path string, args *Args) error {
	if path == "" {
		return nil
	}

	// # Read os env
	viper.AutomaticEnv()

	// # Tell viper the path/location of your config file
	viper.SetConfigFile(path)

	// # Viper reads all the variables from env file and log error if any found
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	config := &Config{}

	// # Viper unmarshals the loaded env varialbes into the struct
	if err := viper.Unmarshal(config); err != nil {
		return err
	}

	fmt.Println(strings.Join(config.Exclude, ","))
	fmt.Println(strings.Join(config.Include, ","))

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