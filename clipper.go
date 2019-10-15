package main

import (
	"log"
	"flag"
	"encoding/json"
	"github.com/munoudesu/clipper/configurator"
	"github.com/munoudesu/clipper/youtubedataapi"
	"github.com/munoudesu/clipper/database"
)

type clipperYoutubeConfig struct {
	ApiKeyFile string `toml:"apiKeyFile"`
	Channels []*youtubedataapi.Channel`toml:"channels"`
}

type clipperDatabaseConfig struct {
	DatabasePath string `toml:"databasePath"`
}

type clipperConfig struct {
	Youtube  *clipperYoutubeConfig    `toml:"youtube"`
	Database *clipperDatabaseConfig   `toml:"database"`
}

func verboseLoadedConfig(verbose bool, loadedConfig *clipperConfig) {
	if !verbose {
		return
	}
	j, err := json.Marshal(loadedConfig)
	if err != nil {
		log.Printf("can not dump config (%v)", err)
		return
	}
	log.Printf("loaded config: %v", string(j))
}

func main() {
	var configFile string
	var verbose bool
	var searchVideo bool
	var searchComment bool
	var recentVideo bool
	var checkVideoModified bool
	var checkCommentModified bool
	flag.StringVar(&configFile, "config", "clipper.conf", "config file")
	flag.BoolVar(&verbose, "verbose", false, "verbose")
	flag.BoolVar(&searchVideo, "searchVideo", true, "search video")
	flag.BoolVar(&searchComment, "searchComment", true, "search comment")
	flag.BoolVar(&recentVideo, "recentVideo", true, "recent video")
	flag.BoolVar(&checkVideoModified, "checkVideoModified", false, "check video modified")
	flag.BoolVar(&checkCommentModified, "checkCommentModified", false, "check comment modified")
	flag.Parse()
	cf, err := configurator.NewConfigurator(configFile)
	conf := new(clipperConfig)
	err = cf.Load(conf)
	if err != nil {
		log.Printf("can not load config: %v", err)
		return
	}
	verboseLoadedConfig(verbose, conf)
	youtubeApiKey, err := youtubedataapi.LoadApiKey(conf.Youtube.ApiKeyFile)
	if err != nil {
		log.Printf("can not load api key of youtube: %v", err)
		return
	}
	databaseOperator, err := database.NewDatabaseOperator(conf.Database.DatabasePath)
	if err != nil {
		 log.Printf("can not create database operator: %v", err)
		 return
	}
	err = databaseOperator.Open()
	if err != nil {
		log.Printf("can not open database: %v", err)
		return
	}
	defer databaseOperator.Close()
	youtubeVideoSearcher, err := youtubedataapi.NewVideoSearcher(youtubeApiKey, conf.Youtube.Channels, databaseOperator)
	if err != nil {
		log.Printf("can not create video searcher of youtube: %v", err)
		return
	}
	err = youtubeVideoSearcher.Search(searchVideo, searchComment, recentVideo, checkVideoModified, checkCommentModified)
	if err != nil {
		log.Printf("can not search youtube video: %v", err)
		return
	}
}
