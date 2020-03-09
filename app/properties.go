package main

import (
	"github.com/artemkaxboy/go-hocon"
	log "github.com/sirupsen/logrus"
)

// Properties struct is used for loading and providing access to configuration file.
type Properties struct {
	Mysql struct {
		URL   string `hocon:"node=url"`
		Table string `hocon:"node=table,default=satellites"`
	} `hocon:"node=mysql"`

	Parser struct {
		BaseURL             string   `hocon:"node=baseUrl"`
		SatelliteURLPattern string   `hocon:"node=satelliteUrlPattern"`
		URLs                []string `hocon:"node=urls"`
	} `hocon:"node=parser"`

	LogLevel string `hocon:"node=logLevel"`
}

var (
	props *Properties
)

// getProperties loads configuration from file to Properties struct if needed and gives pointer to it
func getProperties() *Properties {
	if props == nil {
		props = &Properties{}
		if err := hocon.LoadConfigFile("sat-parser.conf", props); err != nil {
			log.WithError(err).Error("cannot load properties, falling back to example values")
			if err := hocon.LoadConfigFile("sat-parser.conf.example", props); err != nil {
				log.WithError(err).Fatal("cannot load default properties")
			}
		}
	}
	return props
}
