package server

import (
	"fmt"
	"strings"
	"path/filePath"
	"net/http"
	"github.com/gin-gonic/gin"
)

type handler struct {
        cacher *cacher
}

func (h *handler) getCacheFileResponse(c *gin.Context, cacheFilePath string, cacheFileData []byte, contentType string, ok bool) {
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{ "error": fmt.Sprintf("not found %v", cacheFilePath) })
		return
	}
	c.Data(http.StatusOK, contentType, cacheFileData)
}

func (h *handler) getCacheFile(c *gin.Context) {
	acceptEncoding := c.GetHeader("Accept-Encoding")
	gzipOK := strings.Contains(acceptEncoding, "gzip")
	cacheFilePath := c.Param("cacheFilePath")
	if cacheFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{ "error": fmt.Sprintf("not found %v", cacheFilePath) })
	}
	if gzipOK {
		cacheFileData, ok := cacher.getGzip(cacheFilePath)
		c.Header("Content-Encoding", "gzip")
		this.getCacheFileResponse(c, cacheFilePath, cacheFileData, contentType, ok)
	} else {
		cacheFileDataData, ok := cacher.getRaw(cacheFilePath)
		this.getCacheFileResponse(c, cacheFilePath, cacheFileData, contentType, ok)
	}
}

func newHandler(cacher *cacher) (*handler){
	return &handler{
	    cacher: *cacher,
	}
}


