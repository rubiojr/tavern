package server

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rubiojr/tavern/internal/middleware"
)

const UploadRoute = "/v1/tavern/upload"
const ServerDefaultUploadsPath = "tavern_uploads"
const ServerDefaultAddr = "127.0.0.1:8000"
const ServerDefaultURL = "http://" + ServerDefaultAddr
const ServerDefaultCharmServerURL = "https://cloud.charm.sh:35354"

type Config struct {
	Addr                string
	UploadsPath         string
	AllowedCharmServers []string
}

type Server struct {
	config *Config
}

func NewServer() *Server {
	config := &Config{
		Addr:        ServerDefaultAddr,
		UploadsPath: ServerDefaultUploadsPath,
	}

	return NewServerWithConfig(config)
}

func NewServerWithConfig(config *Config) *Server {
	if config.UploadsPath == "" {
		config.UploadsPath = ServerDefaultUploadsPath
	}

	if config.Addr == "" {
		config.Addr = ServerDefaultAddr
	}

	return &Server{config: config}
}

func (s *Server) Serve(ctx context.Context) error {
	err := os.MkdirAll(s.config.UploadsPath, 0755)
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	uploads := router.Group(UploadRoute)
	allowedServers := map[string]struct{}{}
	for _, host := range s.config.AllowedCharmServers {
		allowedServers[host] = struct{}{}
	}
	uploads.Use(middleware.JWKS(allowedServers))
	uploads.POST("/", middleware.Uploads(s.config.UploadsPath, 32<<20))
	router.StaticFS("/", http.Dir(s.config.UploadsPath))
	log.Printf("serving on: %s", s.config.Addr)
	log.Printf("uploads directory: %s", s.config.UploadsPath)

	srv := &http.Server{
		Addr:    s.config.Addr,
		Handler: router,
	}

	go func() {
		srv.ListenAndServe()
	}()

	<-ctx.Done()

	return srv.Shutdown(ctx)
}
