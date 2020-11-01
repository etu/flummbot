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
			Enable      bool
			Separator   string
			UserLogSize int
		}
		Karma struct {
			Enable        bool
			PlusOperator  string
			MinusOperator string
			Command       string
		}
		Quotes struct {
			Enable  bool
			Command string
		}
		Tells struct {
			Enable  bool
			Command string
		}
	}
}

var config ClientConfig

func New(configFile string) ClientConfig {
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

func Get() ClientConfig {
	return config
}
