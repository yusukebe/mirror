package mirror

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type Client struct {
	httpClient   *http.Client
	outputDir    string
	userAgent    string
	mirroredURLs map[string]bool
}

func NewClient(param param) *Client {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	return &Client{
		outputDir:    param.outputDir,
		httpClient:   client,
		userAgent:    param.userAgent,
		mirroredURLs: make(map[string]bool),
	}
}

func (c *Client) mirror(baseURL string) {
	err := c.browse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Client) browse(baseUrl string) error {
	url, err := url.Parse(baseUrl)

	if url.Path == "" {
		url.Path = "/"
	}

	if err != nil {
		log.Fatal(err)
	}

	l, err := launcher.New().
		Set("no-sandbox", "true").
		Launch()

	if err != nil {
		log.Fatal(err)
	}

	browser := rod.New().ControlURL(l).MustConnect()

	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	base := fmt.Sprintf("%s://%s", url.Scheme, url.Hostname())
	wg := &sync.WaitGroup{}

	router.MustAdd(fmt.Sprintf("%s/*", base), func(ctx *rod.Hijack) {
		wg.Add(1)
		requestURL := ctx.Request.URL()

		if c.userAgent != "" {
			ctx.Request.Req().Header.Set("User-Agent", c.userAgent)
		}

		ctx.MustLoadResponse()

		r := regexp.MustCompile(`html`)
		contentType := ctx.Response.Headers().Get("Content-Type")
		if r.MatchString(contentType) {
			wg.Done()
			return
		}

		body, err := c.decodeContent(ctx.Response)
		if err != nil {
			log.Fatal(err)
		}

		var path string
		if requestURL.Path == url.Path {
			path = fmt.Sprintf("%s/index.html", c.outputDir)
		} else {
			path = fmt.Sprintf("%s%s", c.outputDir, requestURL.Path)
		}

		err = c.saveFile(path, []byte(body))
		if err != nil {
			fmt.Println(err)
		}

		c.printSaved(requestURL.String(), path)
		wg.Done()
	})

	go router.Run()
	page := browser.MustPage(url.String())
	page.WaitLoad()
	height := page.MustGetWindow().Height
	// Scroll 100 times
	page.Mouse.Scroll(0, float64(height), 100)
	wg.Wait()
	return nil
}

func (c *Client) printSaved(url string, path string) {
	fmt.Printf("Save %s to %s\n", url, path)
}

func (c *Client) decodeContent(resp *rod.HijackResponse) ([]byte, error) {
	var respBody []byte
	var err error

	if resp.Headers().Get("Content-Encoding") == "br" {
		reader := brotli.NewReader(strings.NewReader(resp.Body()))
		respBody, err = ioutil.ReadAll(reader)
		if err != nil {
			return respBody, err
		}
	} else if resp.Headers().Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(strings.NewReader(resp.Body()))
		if err != nil {
			return respBody, err
		}
		defer reader.Close()
		respBody, err = ioutil.ReadAll(reader)
		if err != nil {
			return respBody, err
		}
	} else {
		respBody = []byte(resp.Body())
	}
	return respBody, nil
}

func (c *Client) saveFile(path string, content []byte) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}
