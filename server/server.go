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
	Addr           string
	UploadsPath    string
	CharmServerURL string
	Whitelist      []string
}

type Server struct {
	config *Config
}

func NewServer() *Server {
	config := &Config{
		Addr:           ServerDefaultAddr,
		UploadsPath:    ServerDefaultUploadsPath,
		CharmServerURL: ServerDefaultCharmServerURL,
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

	envCharmURL := os.Getenv("CHARM_SERVER_URL")
	if config.CharmServerURL == ServerDefaultCharmServerURL && envCharmURL != "" {
		config.CharmServerURL = envCharmURL
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
	whitelist := map[string]struct{}{}
	for _, host := range s.config.Whitelist {
		whitelist[host] = struct{}{}
	}
	uploads.Use(middleware.JWKS(whitelist))
	uploads.POST("/", middleware.Uploads(s.config.UploadsPath, 32<<20))
	router.StaticFS("/", http.Dir(s.config.UploadsPath))
	log.Printf("serving on: %s", s.config.Addr)
	log.Printf("uploads directory: %s", s.config.UploadsPath)
	log.Printf("charm server: %s", s.config.CharmServerURL)

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
