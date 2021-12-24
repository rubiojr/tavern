package client

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/charmbracelet/charm/client"
	cfs "github.com/charmbracelet/charm/fs"
)

const DefaultCharmServerHost = "https://cloud.charm.sh"
const DefaultServerURL = "http://localhost:8000"
const DefaultCharmServerHTTPort = 35354
const DefaultCharmServerSSHPort = 35353

type Client struct {
	remoteFS    *cfs.FS
	charmClient *client.Client
	config      *Config
}

type Config struct {
	ServerURL           string
	CharmServerHost     string
	CharmServerHTTPPort int
	CharmServerSSHPort  int
}

func NewClient() (*Client, error) {
	return NewClientWithConfig(DefaultConfig())
}

func DefaultConfig() *Config {
	return &Config{ServerURL: DefaultServerURL, CharmServerHost: DefaultCharmServerHost, CharmServerHTTPPort: DefaultCharmServerHTTPort, CharmServerSSHPort: DefaultCharmServerSSHPort}
}

func NewClientWithConfig(cfg *Config) (*Client, error) {
	ccfg, err := client.ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	ccfg.Host = cfg.CharmServerHost
	ccfg.HTTPPort = cfg.CharmServerHTTPPort
	ccfg.SSHPort = cfg.CharmServerSSHPort

	c, err := client.NewClient(ccfg)
	if err != nil {
		return nil, err
	}

	remote, err := cfs.NewFSWithClient(c)
	if err != nil {
		return nil, err
	}

	return &Client{config: cfg, remoteFS: remote, charmClient: c}, nil
}

func (c *Client) Publish(path string) error {
	return c.PublishWithRoot("/", path)
}

func (c *Client) PublishWithRoot(root, path string) error {
	var body *bytes.Buffer
	var writer *multipart.Writer

	fmt.Printf("Publishing %s\n", path)
	fmt.Printf("Retrieving files from %s...\n", c.charmClient.Config.Host)
	f, err := c.remoteFS.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("Publishing to %s\n", c.config.ServerURL)
	if info.IsDir() {
		body, writer, err = uploadDir(c.remoteFS, path)
	} else {
		body, writer, err = uploadFile(c.remoteFS, path)
	}
	if err != nil {
		return err
	}

	id, err := c.charmClient.ID()
	if err != nil {
		return err
	}

	req, err := c.authedRequest("/_tavern/upload", id, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		errStatus, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("publishing failed: %s", errStatus)
	}

	fmt.Println("Site published!")
	fmt.Printf("Visit %s/%s\n", c.config.ServerURL, id)
	return nil
}

func (c *Client) authedRequest(path, id string, body *bytes.Buffer) (*http.Request, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf(c.config.ServerURL+"/_tavern/upload"), body)
	if err != nil {
		return nil, err
	}

	jwt, err := c.charmClient.JWT()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", jwt))
	req.Header.Add("CharmId", id)

	return req, nil
}

func uploadDir(cfs fs.FS, root string) (*bytes.Buffer, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	err := fs.WalkDir(cfs, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || d.IsDir() {
			return nil
		}

		publishedPath := strings.TrimPrefix(path, root)
		fmt.Println("Adding ", publishedPath)
		part, err := writer.CreateFormFile("upload[]", publishedPath)
		if err != nil {
			return err
		}

		f, err := cfs.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		out, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		_, err = part.Write(out)
		if err != nil {
			return err
		}

		return nil
	})

	return body, writer, err
}

func uploadFile(remotefs fs.FS, path string) (*bytes.Buffer, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	fmt.Println("Adding ", path)
	part, err := writer.CreateFormFile("upload[]", path)
	if err != nil {
		return nil, nil, err
	}

	f, err := remotefs.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	out, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}

	_, err = part.Write(out)

	return body, writer, err
}
