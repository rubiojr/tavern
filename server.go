package tavern

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const ServerDefaultUploadsPath = "tavern_uploads"
const ServerDefaultAddr = "0.0.0.0:8000"
const ServerDefaultURL = "http://" + ServerDefaultAddr
const charmServerHost = "cloud.charm.sh"
const charmServerPort = "35354"

type Config struct {
	Addr        string
	UploadsPath string
	charmURL    string
}

type Server struct {
	config *Config
}

func (s *Server) upload(c *gin.Context) {
	log.Printf("Upload request received, uploading to %s", s.config.UploadsPath)
	charmID := c.Request.Header.Get("CharmId")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/id/%s", s.config.charmURL, charmID), nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "unexpected error")
		log.Printf("error creating request: %s", err)
		return
	}

	req.Header.Add("Authorization", c.Request.Header.Get("Authorization"))
	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, "communication with %s failed: %s", s.config.charmURL, err)
		return
	}

	if resp.StatusCode != 200 {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	form, _ := c.MultipartForm()
	if form == nil {
		log.Print("no files found")
		c.String(http.StatusBadRequest, "no files uploaded, invalid request")
		return
	}

	files := form.File["upload[]"]
	for _, file := range files {
		_, params, err := mime.ParseMediaType(file.Header.Get("Content-Disposition"))
		dst := filepath.Join(s.config.UploadsPath, charmID, "/", params["filename"])
		dstDir := filepath.Dir(filepath.Clean(dst))
		os.MkdirAll(dstDir, 0755)
		d := filepath.Clean(filepath.Dir(file.Filename))
		os.MkdirAll(d, 0750)
		err = c.SaveUploadedFile(file, dst)
		if err != nil {
			log.Printf("error saving file: %s", err)
		}
	}
	c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
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

	if config.charmURL == "" {
		host := charmServerHost
		port := charmServerPort
		if os.Getenv("CHARM_HOST") != "" {
			host = os.Getenv("CHARM_HOST")
		}
		if os.Getenv("CHARM_PORT") != "" {
			port = os.Getenv("CHARM_PORT")
		}
		config.charmURL = fmt.Sprintf("https://%s:%s", host, port)
	}

	return &Server{config: config}
}

func (s *Server) Serve() error {
	err := os.MkdirAll(s.config.UploadsPath, 0755)
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/_tavern/upload", s.upload)
	router.StaticFS("/", http.Dir(s.config.UploadsPath))
	log.Printf("serving on: %s", s.config.Addr)
	log.Printf("uploads directory: %s", s.config.UploadsPath)
	log.Printf("charm server: %s", s.config.charmURL)

	return http.ListenAndServe(s.config.Addr, router)
}
