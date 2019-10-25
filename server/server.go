package server

import (
        "log"
        "context"
        "time"
        "net/http"
        "github.com/gin-gonic/gin"
)

type server struct {
	cacher       *cacher
	addrPort     string
        tlsCertPath  string
        tlsKeyPath   string
	rootDirPath  string
	cacheDirPath string
	release      bool
        server       *http.Server
        router       *gin.Engine
	handler      *handler
	verbose      bool
}

func (s *server) Start() {
	s.cacher.start()
        go func() {
                err := s.server.ListenAndServeTLS(s.tlsCertPath, s.tlsKeyPath);
                if err != nil && err != http.ErrServerClosed {
                        log.Fatalf("listen: %v", err)
                }
        }()
}

func (s *server) Stop() {
        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
        defer cancel()
        err := s.server.Shutdown(ctx)
        if err != nil {
            log.Printf("Server Shutdown: %v", err)
        }
	s.cacher.stop()
}

func NewServer(addrPort string, tlsCertPath string, tlsKeyPath string, rootDirPath string, cacheDirPath string, release bool, verbose bool) (*server) {
	if release {
		gin.SetMode(gin.ReleaseMode)
	}
        cacher := newCacher(cacheDirPath, verbose)
        handler := newHandler(cacher, verbose)
        router := gin.Default()
        router.Static("/root", rootDirPath)
        jsonRouter := router.Group("/cache")
        jsonRouter.HEAD("/:cacheFilePath", handler.headCacheFile)
        jsonRouter.GET("/:cacheFilePath", handler.getCacheFile)
        s := &http.Server{
            Addr: addrPort,
            Handler: router,
            IdleTimeout: time.Duration(30) * time.Second,
        }
        return &server {
		cacher: cacher,
		addrPort: addrPort,
		tlsCertPath: tlsCertPath,
		tlsKeyPath: tlsKeyPath,
		rootDirPath: rootDirPath,
		cacheDirPath: cacheDirPath,
		release: release,
		server: s,
		router: router,
		handler: handler,
		verbose: verbose,
        }
}

