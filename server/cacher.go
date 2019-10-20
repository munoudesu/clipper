package server

import (
	"log"
	"time"
	"mime"
	"sync"
	"path/filepath"
	"compress/gzip"
)

type fileCache struct {
	modifiedTime  time.Time
	sha1Sigest    []byte
	gzipData      []byte
	rawData       []byte
	mineType      string
}

type cacher struct {
	cacheDirPath                string
	mutex                       *sync.Mutex
	fileCache                   map[string]fileCache
	cacheLoopFinishResquestChan chan int
	cacheLoopFinishResponseChan chan int
}

func (c *cacher) update(dataFilePath string, digestFilePath string, dataFile os.FileInfo, digestFile os.FileInfo) (error) [

	// XXXXXXXXXXXXXXxx
	fileCache, ok := c.fileCache[dataFilePath]
	if !ok {

		func TypeByExtension(ext string) string
		err = ioutil.ReadFile(channelPageJsonSha1DigestPath, []byte(newChannelPageSha1Digest), 0644)
                if err != nil {
                        return errors.Wrapf(err, "can not write sha1 digest of json to file (channelId = %v, path = %v)", dbChannel.ChannelId, channelPageJsonSha1DigestPath)
                }

		newFileCache = &fileCache{
			modifiedTime: digestFile.ModTime(),
		}


	}
}

func (c *cacher) dirWalk(dir string) (error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		errors.Wrapf(err, "can not readdir (%v)", dir)
	}
	for _, findFile := range files {
		findFilePath := filepath.Join(dir, findFile.Name())
		if findFile.IsDir() {
			c.dirWalk(findFilePath)
			continue
		}
		ext := filepath.Ext(findFilePath)
		if ext == ".sha1" {
			dataFilePath := strings.TrimSuffix(find, ext)
			dataFile, err := os.Stat(dataFilePath)
			if err != nil {
				log.Printf("skip cache, not found data file (data file = %v, digest file = %v)", dataPath, findFilePath)
				continue
			}
			err := c.update(dataPath, findFilePath, dataFile, findFile)
			if err != nil {
				log.Printf("skip cache, can not update cache (data file = %v, digest file = %v)", dataPath, findFilePath)
				continue
			}
		}
		continue
	}
}

func (c *cacher) cacheMain() {
	c.dirWalk(c.cacheDirPath)
}

func (c *cacher) cacheLoop() {
        c.cacheMain(ex, notifier)
        for {
                select {
                case <-time.After(time.Second):
                        c.cacheMain()
                case <-c.cacheLoopFinishResquestChan:
                        goto LAST
                }
        }
LAST:
        close(c.cacheLoopFinishResponseChan)
}

func (c *cacher) Start() {
        go c.cacheLoop()
}

func (c *cacher) Stop() {
        close(c.-c.cacheLoopFinishResquestChan)
        <-c.cacheLoopFinishResponseChan
}

func newCacher(cacheDirPath string) (*cacher) {
        return &cacher {
		cacheDirPath: cacheDirPath,
                mutex: new(sync.Mutex),
                fileCache: make(map[string]fileCache),
                cacheLoopFinishResquestChan: make(chan int),
                cacheLoopFinishResponseChan: make(chan int),
        }
}

