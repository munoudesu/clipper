package main

import (
	"os"
	"log"
	"flag"
	"syscall"
	"os/signal"
	"encoding/json"
	"github.com/munoudesu/clipper/configurator"
	"github.com/munoudesu/clipper/youtubedataapi"
	"github.com/munoudesu/clipper/twitterapi"
	"github.com/munoudesu/clipper/database"
	"github.com/munoudesu/clipper/builder"
	"github.com/munoudesu/clipper/server"
)

type clipperYoutubeConfig struct {
	ApiKeyFile string                  `toml:"apiKeyFile"`
	MaxVideos  int64                   `toml:"maxVideos"`
	Channels   youtubedataapi.Channels `toml:"channels"`
}

type clipperTwitterConfig struct {
	ApiKeyFile string      `toml:"apiKeyFile"`
	Users twitterapi.Users `toml:"users"`
}

type clipperDatabaseConfig struct {
	DatabasePath string `toml:"databasePath"`
}

type clipperBuilderConfig struct {
	SourceDirPath       string `toml:"sourceDirPath"`
	BuildDirPath        string `toml:"buildDirPath"`
	MaxDuration         int64  `toml:"maxDuration"`
	AdjustStartTimeSpan int64  `toml:"adjustStartTimeSpan"`
}

type clipperWebConfig struct {
	AddrPort     string `toml:"addrPort"`
	TlsCertPath  string `toml:"tlsCertPath"`
	TlsKeyPath   string `toml:"tlsKeyPath"`
	RootDirPath  string `toml:"rootDirPath"`
	CacheDirPath string `toml:"cacheDirPath"`
}

type clipperIpfsConfig struct {
	AddrPort string `toml:"addrPort"`
}

type clipperConfig struct {
	Youtube  *clipperYoutubeConfig  `toml:"youtube"`
	Twitter  *clipperTwitterConfig  `toml:"twitter"`
	Database *clipperDatabaseConfig `toml:"database"`
	Builder  *clipperBuilderConfig  `toml:"builder"`
	Web      *clipperWebConfig      `toml:"web"`
	Ipfs     *clipperIpfsConfig     `toml:"ipfs"`
}

type commandArguments struct {
	configFile           string
	verbose              bool
	searchChannel        bool
	searchVideo          bool
	searchComment        bool
	checkChannelModified bool
	checkVideoModified   bool
	checkCommentModified bool
	skipSearch           bool
	skipBuild            bool
	skipNotify           bool
	runMode              string
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

func crawl(conf *clipperConfig, cmdArgs *commandArguments) {
	youtubeApiKeys, err := youtubedataapi.LoadApiKey(conf.Youtube.ApiKeyFile)
	if err != nil {
		log.Printf("can not load api key of youtube: %v", err)
		os.Exit(1)
	}
	twitterApiKeyPair, err := twitterapi.LoadApiKey(conf.Twitter.ApiKeyFile)
	if err != nil {
		log.Printf("can not load api key pair of twitter: %v", err)
		os.Exit(1)
	}
	databaseOperator, err := database.NewDatabaseOperator(conf.Database.DatabasePath)
	if err != nil {
		log.Printf("can not create database operator: %v", err)
		os.Exit(1)
	}
	err = databaseOperator.Open()
	if err != nil {
		log.Printf("can not open database: %v", err)
		os.Exit(1)
	}
	defer databaseOperator.Close()
	if !cmdArgs.skipSearch {
		youtubeSearcher, err := youtubedataapi.NewSearcher(youtubeApiKeys, conf.Youtube.MaxVideos, conf.Youtube.Channels, databaseOperator)
		if err != nil {
			log.Printf("can not create searcher of youtube: %v", err)
			os.Exit(1)
		}
		err = youtubeSearcher.Search(cmdArgs.searchChannel, cmdArgs.searchVideo, cmdArgs.searchComment, cmdArgs.checkChannelModified, cmdArgs.checkVideoModified, cmdArgs.checkCommentModified)
		if err != nil {
			log.Printf("can not search youtube: %v", err)
			os.Exit(1)
		}
	}
	if !cmdArgs.skipBuild {
		builder, err := builder.NewBuilder(conf.Builder.SourceDirPath, conf.Builder.BuildDirPath, conf.Builder.MaxDuration, conf.Builder.AdjustStartTimeSpan, conf.Youtube.Channels, databaseOperator)
		if err != nil {
			log.Printf("can not create builder: %v", err)
			os.Exit(1)
		}
		err = builder.Build()
		if err != nil {
			log.Printf("can not build page: %v", err)
			os.Exit(1)
		}
	}
	log.Printf("%v",twitterApiKeyPair)
}

func signalWait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGQUIT:
			fallthrough
		case syscall.SIGTERM:
			return
		default:
			log.Printf("unexpected signal (sig = %v)", sig)
		}
	}
}

func web(conf *clipperConfig, cmdArgs *commandArguments) {
	server := server.NewServer(conf.Web.AddrPort, conf.Web.TlsCertPath, conf.Web.TlsKeyPath, conf.Web.RootDirPath, conf.Web.CacheDirPath)
	server.Start()
	signalWait()
	server.Stop()
}

func main() {
	cmdArgs := new(commandArguments)
	flag.StringVar(&cmdArgs.configFile, "config", "clipper.conf", "config file")
	flag.BoolVar(&cmdArgs.verbose, "verbose", false, "verbose")
	flag.BoolVar(&cmdArgs.searchChannel, "searchChannel", true, "search channel")
	flag.BoolVar(&cmdArgs.searchVideo, "searchVideo", true, "search video")
	flag.BoolVar(&cmdArgs.searchComment, "searchComment", true, "search comment")
	flag.BoolVar(&cmdArgs.checkChannelModified, "checkChannelModified", true, "check channel modified")
	flag.BoolVar(&cmdArgs.checkVideoModified, "checkVideoModified", true, "check video modified")
	flag.BoolVar(&cmdArgs.checkCommentModified, "checkCommentModified", true, "check comment modified")
	flag.BoolVar(&cmdArgs.skipSearch, "skipSearch", false, "skip search")
	flag.BoolVar(&cmdArgs.skipBuild, "skipBuild", false, "skip build")
	flag.BoolVar(&cmdArgs.skipNotify, "skipNotify", false, "skip Notify")
	flag.StringVar(&cmdArgs.runMode, "runMode", "crawl", "run mode crawl or web")
	flag.Parse()
	cf, err := configurator.NewConfigurator(cmdArgs.configFile)
	conf := new(clipperConfig)
	err = cf.Load(conf)
	if err != nil {
		log.Printf("can not load config: %v", err)
		os.Exit(1)
	}
	verboseLoadedConfig(cmdArgs.verbose, conf)
	if cmdArgs.runMode == "crawl" {
		crawl(conf, cmdArgs)
	} else if cmdArgs.runMode == "web" {
		web(conf, cmdArgs)
	}
}
