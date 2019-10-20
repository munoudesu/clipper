package server

import (
        "log"
        "context"
        "time"
        "path"
        "path/filepath"
        "net/http"
        "github.com/pkg/errors"
        "github.com/gin-contrib/sessions"
        "github.com/gin-contrib/sessions/cookie"
        "github.com/gin-gonic/gin"
        "github.com/potix/ensemble/common/db"
)

type server struct {
	cacher       *cacher
	addrPort     string
        tlsCertPath  string
        tlsKeyPath   string
	rootDirPath  string
	cacheDir     string
        server       *http.Server
        router       *gin.Engine
	hander       *handler
}

func (s *server) start() {
	s.cacher.Start()
        go func() {
                err := s.server.ListenAndServeTLS(s.tlsCert, s.tlsKey);
                if err != nil && err != http.ErrServerClosed {
                        log.Fatalf("listen: %v", err)
                }
        }()
}

func (s *server) stop() {
        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
        defer cancel()
        err := s.server.Shutdown(ctx)
        if err != nil {
            log.Printf("Server Shutdown: %v", err)
        }
	s.cacher.Stop()
}

func newServer(conf *config, addrPort string, tlsCertPath string, tlsKeyPath string, rootDirPath string, cacheDir string) (*server, error) {
	cacheDirPath := filePath.Join(rootDirPath, cacheDirPath)
        cacher, err := newCacher(cacheDirPath)
        if err != nil {
                return nil, errors.Wrap(err, "can not create cacher")
        }
        handler, err := newHandler(cacher)
        if err != nil {
                return nil, errors.Wrap(err, "can not create handler")
        }
        router := gin.Default()
        router.Static("/", buildDirPath)
        jsonRouter := router.Group(path.Join("/", cacheDirPath))
        jsonRouter.GET("/:cacheFilePath", handler.getJson)
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
		cacheDir: cacheDir,
		server: s,
		router: router,
		handler: handler,
        }, nil
}

