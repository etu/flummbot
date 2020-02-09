package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
)

type ClientConfig struct {
	Connections []struct {
		Name             string
		Server           string
		Port             uint16
		Channels         []string
		User             string
		Nick             string
		Password         string
		UseTLS           bool
		NickservIdentify string
	}
	Database struct {
		Dialect string
		Args    string
	}
	Modules struct {
		Corrections struct {
			Separator string
		}
		Karma struct {
			PlusOperator  string
			MinusOperator string
			Command       string
		}
		Quotes struct {
			Command string
		}
		Tells struct {
			Command string
		}
	}
}

func New(configFile string) ClientConfig {
	var config ClientConfig

	// Read the config file
	fileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("File read error: %v", err))
	}

	// Parse config
	if _, err := toml.Decode(string(fileContent), &config); err != nil {
		log.Fatal(fmt.Sprintf("Config error: %v", err))
	}

	return config
}
