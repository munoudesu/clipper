package main

import (
	"log"
	"flag"
	"encoding/json"
	"github.com/munoudesu/clipper/configurator"
	"github.com/munoudesu/clipper/youtubedataapi"
)

type clipperYoutubeConfig struct {
	ApiKeyFile string `toml:"apiKeyFile"`
}

type clipperDatabaseConfig struct {
	DatabasePath string `toml:"databasePath"`
}

type clipperChannelConfig struct {
	Name      string `toml:"name"`
	ChannelId string `toml:"channelId"`
}

type clipperConfig struct {
	Youtube  *clipperYoutubeConfig   `toml:"youtube"`
	Database *clipperDatabaseConfig  `toml:"database"`
	Channels []*clipperChannelConfig `toml:"channels"`
}

func verboseLoadedConfig(verbose bool, loadedConfig *clipperConfig) {
	if !verbose {
		return
	}
	j, err := json.Marshal(loadedConfig)
	if err != nil {
		log.Fatalf("can not dump config (%v)", err)
	}
	log.Printf("loaded config: %v", string(j))
}

func main() {
	var configFile string
	var verbose bool
	flag.StringVar(&configFile, "config", "clipper.conf", "config file")
	flag.BoolVar(&verbose, "verbose", false, "verbose")
	flag.Parse()
	cf, err := configurator.NewConfigurator(configFile)
	conf := new(clipperConfig)
	err = cf.Load(conf)
	if err != nil {
		log.Fatalf("can not load config: %v", err)
	}
	verboseLoadedConfig(verbose, conf)
	youtubeApiKey, err := youtubedataapi.LoadApiKey(conf.Youtube.ApiKeyFile)
	if err != nil {
		log.Fatalf("can not load api key of youtube: %v", err)
	}
	log.Printf("%v", youtubeApiKey)
}
