package server

import (
	"os"
	"log"
	"time"
	"mime"
	"sync"
	"strings"
	"bytes"
	"path/filepath"
	"io/ioutil"
	"compress/gzip"
	"github.com/pkg/errors"
)

type fileCache struct {
	digestFilModTime  time.Time
	dataFilModTime    time.Time
	sha1Digest        []byte
	rawData           []byte
	gzipData          []byte
	mineType          string
}

type cacher struct {
	cacheDirPath                string
	mutex                       *sync.Mutex
	filesCache                  map[string]*fileCache
	cacheLoopFinishResquestChan chan int
	cacheLoopFinishResponseChan chan int
	verbose                     bool
}

func (c *cacher) getGzipData(urlDataFilePath string) ([]byte, string, bool) {
	dataFilePath := filepath.Join(c.cacheDirPath, urlDataFilePath)
	fileCache, ok := c.getFileCache(dataFilePath)
	if !ok {
		return nil, "", false
	}
	return fileCache.gzipData, fileCache.mineType, ok
}

func (c *cacher) getRawData(urlDataFilePath string) ([]byte, string, bool) {
	dataFilePath := filepath.Join(c.cacheDirPath, urlDataFilePath)
	fileCache, ok := c.getFileCache(dataFilePath)
	if !ok {
		return nil, "", false
	}
	return fileCache.rawData, fileCache.mineType, ok
}

func (c *cacher) getFileCache(dataFilePath string) (*fileCache, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	fileCache, ok := c.filesCache[dataFilePath]
	if c.verbose && !ok {
		log.Printf("not found file cache (%v)", dataFilePath)
	}
	return fileCache, ok
}

func (c *cacher) setFileCache(dataFilePath string, newFileCache *fileCache) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.verbose {
		log.Printf("set file cache (%v)", dataFilePath)
	}
	c.filesCache[dataFilePath] = newFileCache
}

func (c *cacher) gzipCompress(rawData []byte) ([]byte, error) {
	byteBuffer := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(byteBuffer)
	_, err := gzipWriter.Write(rawData);
	if err != nil {
		return nil, errors.Wrap(err, "can not write to gzip writer")
	}
	gzipWriter.Flush()
	gzipWriter.Close()
	return byteBuffer.Bytes(), nil
}

func (c *cacher) updateMain(dataFilePath string, digestFilePath string, dataFile os.FileInfo, digestFile os.FileInfo) (error) {
	findFileCache, ok := c.getFileCache(dataFilePath)
	if !ok {
		sha1Digest, err := ioutil.ReadFile(digestFilePath)
		if err != nil {
			return errors.Wrapf(err, "can not read digest file (%v)", digestFilePath)
		}
		rawData, err := ioutil.ReadFile(dataFilePath)
		if err != nil {
			return errors.Wrapf(err, "can not read data file (%v)", dataFilePath)
		}
		gzipData, err := c.gzipCompress(rawData)
		if err != nil {
			return errors.Wrapf(err, "can not compress raw data (%v)", dataFilePath)
		}
		dataFileExt := filepath.Ext(dataFilePath)
		mineType := mime.TypeByExtension(dataFileExt)
		newFileCache := &fileCache {
			digestFilModTime: digestFile.ModTime(),
			dataFilModTime: digestFile.ModTime(),
			sha1Digest: sha1Digest,
			rawData: rawData,
			gzipData: gzipData,
			mineType: mineType,
		}
		c.setFileCache(dataFilePath, newFileCache)
		return nil
	}
	if findFileCache.digestFilModTime == digestFile.ModTime() && findFileCache.dataFilModTime == dataFile.ModTime() {
		if c.verbose {
			log.Printf("not changed modified timestamp (data file = %v, digest file = %v)", dataFilePath, digestFilePath)
		}
		return nil
	}
	newSha1Digest, err := ioutil.ReadFile(digestFilePath)
	if err != nil {
		return errors.Wrapf(err, "can not read digest file (%v)", digestFilePath)
	}
	if bytes.Compare(findFileCache.sha1Digest, newSha1Digest) == 0 {
		if c.verbose {
			log.Printf("not changed digest (data file = %v, digest file = %v)", dataFilePath, digestFilePath)
		}
		return nil
	}
	newRawData, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		return errors.Wrapf(err, "can not read data file (%v)", dataFilePath)
	}
	newGzipData, err := c.gzipCompress(newRawData)
	if err != nil {
		return errors.Wrapf(err, "can not compress raw data (%v)", dataFilePath)
	}
	dataFileExt := filepath.Ext(dataFilePath)
	newMineType := mime.TypeByExtension(dataFileExt)
	newFileCache := &fileCache {
		digestFilModTime: digestFile.ModTime(),
		dataFilModTime: dataFile.ModTime(),
		sha1Digest: newSha1Digest,
		rawData: newRawData,
		gzipData: newGzipData,
		mineType: newMineType,
	}
	c.setFileCache(dataFilePath, newFileCache)
	return nil
}

func (c *cacher) dirWalk(dir string) (error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.Wrapf(err, "can not readdir (%v)", dir)
	}
	for _, findFile := range files {
		findFilePath := filepath.Join(dir, findFile.Name())
		if findFile.IsDir() {
			err := c.dirWalk(findFilePath)
			if err != nil {
				if c.verbose {
					log.Printf("can not walk directory (dir = %v): %v", dir, err)
				}
			}
			continue
		}
		ext := filepath.Ext(findFilePath)
		if ext == ".sha1" {
			dataFilePath := strings.TrimSuffix(findFilePath, ext)
			dataFile, err := os.Stat(dataFilePath)
			if err != nil {
				if c.verbose {
					log.Printf("skip cache, not found data file (data file = %v, digest file = %v): %v", dataFilePath, findFilePath, err)
				}
				continue
			}
			err = c.updateMain(dataFilePath, findFilePath, dataFile, findFile)
			if err != nil {
				if c.verbose {
					log.Printf("skip cache, can not update cache (data file = %v, digest file = %v): %v", dataFilePath, findFilePath, err)
				}
				continue
			}
		}
	}
	return nil
}

func (c *cacher) cacheMain() {
	if c.verbose {
		log.Printf("cache process start (cache dir = %v)", c.cacheDirPath)
	}
	err := c.dirWalk(c.cacheDirPath)
	if err != nil {
		log.Printf("can not walk directory (cache dir = %v): %v", c.cacheDirPath, err)
	}
}

func (c *cacher) cacheLoop() {
        c.cacheMain()
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

func (c *cacher) start() {
        go c.cacheLoop()
}

func (c *cacher) stop() {
        close(c.cacheLoopFinishResquestChan)
        <-c.cacheLoopFinishResponseChan
}

func newCacher(cacheDirPath string, verbose bool) (*cacher) {
        return &cacher {
		cacheDirPath: cacheDirPath,
                mutex: new(sync.Mutex),
                filesCache: make(map[string]*fileCache),
                cacheLoopFinishResquestChan: make(chan int),
                cacheLoopFinishResponseChan: make(chan int),
		verbose: verbose,
        }
}

