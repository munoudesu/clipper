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

func (h *handler) containsGzip(acceptEncoding string) (bool) {
	for _, s:= range strings.Split(acceptEncoding, ",") {
		if "gzip" == strings.TrimSpace(s) {
			return true
		}
	}
	return false
}

func (h *handler) getCacheFile(c *gin.Context) {
	gzipOK := h.containsGzip(c.GetHeader("Accept-Encoding"))
	ifNotMatch := c.GetHeader("If-None-Match")
	ifModifiedSince:= c.GetHeader("If-Modified-Since")
	cacheFilePath := c.Param("cacheFilePath")
	cacheFileRawData, cacheFileGzipData, contentType, etag, lastModified, modified, ok := h.cacher.getData(cacheFilePath, ifNotMatch, ifModifiedSince)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{ "error": fmt.Sprintf("not found %v", cacheFilePath) })
		return
	}
	c.Header("Etag", etag)
	if ifNotMatch != "" {
		if modified == false {
			c.Status(http.StatusNotModified)
			return
		}
		if gzipOK {
			c.Header("Content-Encoding", "gzip")
			c.Data(http.StatusOK, contentType, cacheFileGzipData)
			return
		}
		c.Data(http.StatusOK, contentType, cacheFileRawData)
		return
	}
	if ifModifiedSince != "" {
		c.Header("Last-Modified", lastModified.Format(http.TimeFormat))
		if modified == false {
			c.Status(http.StatusNotModified)
			return
		}
		if gzipOK {
			c.Header("Content-Encoding", "gzip")
			c.Data(http.StatusOK, contentType, cacheFileGzipData)
			return
		}
		c.Data(http.StatusOK, contentType, cacheFileRawData)
		return
	}
	if gzipOK {
		c.Header("Content-Encoding", "gzip")
		c.Data(http.StatusOK, contentType, cacheFileGzipData)
		return
	}
	c.Data(http.StatusOK, contentType, cacheFileRawData)
	return
}

func newHandler(cacher *cacher, verbose bool) (*handler){
	return &handler{
	    cacher: cacher,
	    verbose: verbose,
	}
}

