package server

import (
	"fmt"
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
)

type handler struct {
        cacher  *cacher
	verbose bool
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
		cacheFileData, contentType, ok := h.cacher.getGzipData(cacheFilePath)
		c.Header("Content-Encoding", "gzip")
		h.getCacheFileResponse(c, cacheFilePath, cacheFileData, contentType, ok)
	} else {
		cacheFileData, contentType, ok := h.cacher.getRawData(cacheFilePath)
		h.getCacheFileResponse(c, cacheFilePath, cacheFileData, contentType, ok)
	}
}

func newHandler(cacher *cacher, verbose bool) (*handler){
	return &handler{
	    cacher: cacher,
	    verbose: verbose,
	}
}

