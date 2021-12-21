package tavern

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/charm/client"
	cfs "github.com/charmbracelet/charm/fs"
)

type Client struct {
	remoteFS    *cfs.FS
	charmClient *client.Client
	config      *ClientConfig
}

type ClientConfig struct {
	serverURL string
}

func NewClient(serverURL string) (*Client, error) {
	c, err := client.NewClientWithDefaults()
	if err != nil {
		return nil, err
	}

	remote, err := cfs.NewFSWithClient(c)
	if err != nil {
		return nil, err
	}

	config := &ClientConfig{serverURL: serverURL}

	return &Client{config: config, remoteFS: remote, charmClient: c}, nil
}

func (c *Client) Publish(path string) error {
	return c.PublishWithRoot("/", path)
}

func (c *Client) PublishWithRoot(root, path string) error {
	var body *bytes.Buffer
	var writer *multipart.Writer

	f, err := c.remoteFS.Open("/test")
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("Publishing to %s\n", c.config.serverURL)
	if info.IsDir() {
		body, writer = uploadDir(c.remoteFS, path)
	} else {
		body, writer = uploadFile(c.remoteFS, path)
	}

	id, err := c.charmClient.ID()
	if err != nil {
		return err
	}

	req, err := c.authedRequest("/_tavern/upload", id, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error publishing! server error: %s", resp.Status)
	}

	fmt.Println("Site published!")
	fmt.Printf("Visit %s/%s\n", c.config.serverURL, id)
	return nil
}

func (c *Client) authedRequest(path, id string, body *bytes.Buffer) (*http.Request, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf(c.config.serverURL+"/_tavern/upload"), body)
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

func uploadDir(cfs fs.FS, root string) (*bytes.Buffer, *multipart.Writer) {
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

	if err != nil {
		panic(err)
	}

	return body, writer
}

func uploadFile(remotefs fs.FS, path string) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	publishedPath := filepath.Base(path)
	fmt.Println("adding ", publishedPath)
	part, _ := writer.CreateFormFile("upload[]", publishedPath)
	f, err := remotefs.Open(path)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(part, f)
	if err != nil {
		panic(err)
	}

	fmt.Println(body.String())

	return body, writer
}
