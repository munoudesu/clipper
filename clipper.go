package main

import (
	"log"
	"flag"
	"encoding/json"
	"github.com/munoudesu/clipper/configurator"
	"github.com/munoudesu/clipper/youtubedataapi"
	"github.com/munoudesu/clipper/twitterapi"
	"github.com/munoudesu/clipper/database"
	"github.com/munoudesu/clipper/builder"
)

type clipperYoutubeConfig struct {
	ApiKeyFile string                `toml:"apiKeyFile"`
	Channels youtubedataapi.Channels `toml:"channels"`
}

type clipperTwitterConfig struct {
	ApiKeyFile string      `toml:"apiKeyFile"`
	Users twitterapi.Users `toml:"users"`
}

type clipperDatabaseConfig struct {
	DatabasePath string `toml:"databasePath"`
}

type clipperBuilderConfig struct {
	BuildDirPath          string `toml:"buildDirPath"`
	TemplateDirPath       string `toml:"templateDirPath"`
	MaxDuration           int64  `toml:"maxDuration"`
	AdjustStartTimeSpan   int64  `toml:"adjustStartTimeSpan"`
}

type clipperIpfsConfig struct {
	AddrPort string `toml:"addrPort"`
}

type clipperConfig struct {
	Youtube  *clipperYoutubeConfig  `toml:"youtube"`
	Twitter  *clipperTwitterConfig  `toml:"twitter"`
	Database *clipperDatabaseConfig `toml:"database"`
	Builder  *clipperBuilderConfig  `toml:"builder"`
	Ipfs     *clipperIpfsConfig     `toml:"ipfs"`
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
	var skipSearch bool
	var skipBuild bool
	var skipPublish bool
	flag.StringVar(&configFile, "config", "clipper.conf", "config file")
	flag.BoolVar(&verbose, "verbose", false, "verbose")
	flag.BoolVar(&searchVideo, "searchVideo", true, "search video")
	flag.BoolVar(&searchComment, "searchComment", true, "search comment")
	flag.BoolVar(&recentVideo, "recentVideo", true, "recent video")
	flag.BoolVar(&checkVideoModified, "checkVideoModified", false, "check video modified")
	flag.BoolVar(&checkCommentModified, "checkCommentModified", false, "check comment modified")
	flag.BoolVar(&skipSearch, "skipSearch", false, "skip search")
	flag.BoolVar(&skipBuild, "skipBuild", false, "skip build")
	flag.BoolVar(&skipPublish, "skipPublish", false, "skip Publish")
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
	twitterApiKeyPair, err := twitterapi.LoadApiKey(conf.Twitter.ApiKeyFile)
	if err != nil {
		log.Printf("can not load api key pair of twitter: %v", err)
		return
	}
	log.Printf("api key = %v api secret key = %v", twitterApiKeyPair.ApiKey, twitterApiKeyPair.ApiSecretKey)
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
	if !skipSearch {
		youtubeSearcher, err := youtubedataapi.NewSearcher(youtubeApiKey, conf.Youtube.Channels, databaseOperator)
		if err != nil {
			log.Printf("can not create searcher of youtube: %v", err)
			return
		}
		err = youtubeSearcher.Search(searchVideo, searchComment, recentVideo, checkVideoModified, checkCommentModified)
		if err != nil {
			log.Printf("can not search youtube: %v", err)
			return
		}
	}
	if !skipBuild {
		builder, err := builder.NewBuilder(conf.Builder.BuildDirPath, conf.Builder.TemplateDirPath, conf.Builder.MaxDuration, conf.Builder.AdjustStartTimeSpan, conf.Youtube.Channels, databaseOperator)
		if err != nil {
			log.Printf("can not create builder: %v", err)
			return
		}
		err = builder.Build()
		if err != nil {
			log.Printf("can not build page: %v", err)
			return
		}
	}
}
