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

	"github.com/andybalholm/brotli"
)

type Client struct {
	httpClient   *http.Client
	outputDir    string
	userAgent    string
	mirroredURLs map[string]bool
	baseURL      string
}

func NewClient(param param) *Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	return &Client{
		outputDir:    param.outputDir,
		httpClient:   client,
		userAgent:    param.userAgent,
		mirroredURLs: make(map[string]bool),
	}
}

func (c *Client) mirror(baseURL string) {
	c.baseURL = baseURL

	body, err := c.getContent(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	path := fmt.Sprintf("%s/index.html", c.outputDir)

	err = c.saveFile(path, body)
	if err != nil {
		log.Fatal(err)
	} else {
		c.printSaved(baseURL, path)
	}

	bodyString := string(body)
	var links []string

	linkHref, err := c.getLinksFromHTML(bodyString, "href")
	if err != nil {
		log.Fatal(err)
	}
	linkSrc, err := c.getLinksFromHTML(bodyString, "src")
	if err != nil {
		log.Fatal(err)
	}
	linkCSS, err := c.getLinksFromCSS(bodyString)
	if err != nil {
		log.Fatal(err)
	}
	links = append(linkHref, linkSrc...)
	links = append(links, linkCSS...)

	r := regexp.MustCompile(`\.css$`)

	for _, link := range links {
		c.saveLink(link)
		if r.MatchString(link) {
			err = c.mirrorCSS(link)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (c *Client) mirrorCSS(path string) error {
	url := c.getURLFromPath(path)
	body, err := c.getContent(url)
	if err != nil {
		return err
	}
	links, err := c.getLinksFromCSS(string(body))
	if err != nil {
		return err
	}
	for _, link := range links {
		c.saveLink(link)
	}
	return nil
}

func (c *Client) saveLink(path string) error {
	url := c.getURLFromPath(path)
	if c.mirroredURLs[url] {
		//fmt.Printf("%s is already mirrored\n", url)
		return nil
	}

	path = fmt.Sprintf("%s%s", c.outputDir, path)

	body, err := c.getContent(url)

	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		err = c.saveFile(path, body)
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {
			c.mirroredURLs[url] = true
			c.printSaved(url, path)
		}
	}
	return nil
}

func (c *Client) getURLFromPath(path string) string {
	u, _ := c.parseURL(c.baseURL)
	result := fmt.Sprintf("%s://%s%s", u.Scheme, u.Hostname(), path)
	return result
}

func (c *Client) printSaved(url string, path string) {
	fmt.Printf("Save %s to %s\n", url, path)
}

func (c *Client) getContent(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := c.decodeContent(resp)
	if err != nil {
		return body, err
	}
	fmt.Printf("%s - %s\n", resp.Status, url)
	return body, nil
}

func (c *Client) decodeContent(resp *http.Response) ([]byte, error) {
	var respBody []byte
	var err error

	if resp.Header.Get("Content-Encoding") == "br" {
		reader := brotli.NewReader(resp.Body)
		respBody, err = ioutil.ReadAll(reader)
		if err != nil {
			return respBody, err
		}
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return respBody, err
		}
		defer reader.Close()
		respBody, err = ioutil.ReadAll(reader)
		if err != nil {
			return respBody, err
		}
	} else {
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return respBody, err
		}
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

func (c *Client) parseURL(inputURL string) (*url.URL, error) {
	u, err := url.Parse(inputURL)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (c *Client) getLinksFromHTML(html string, attr string) ([]string, error) {
	html = c.chop(html)
	reg := fmt.Sprintf("(?m)<(img|link|script).+?%s=\"(/[^/][^>]+?)\"", attr)
	r := regexp.MustCompile(reg)
	results := r.FindAllStringSubmatch(html, -1)
	var links []string
	r = regexp.MustCompile(".+/$")
	for _, result := range results {
		path := result[2]
		u, _ := c.parseURL(path)
		if r.MatchString(u.Path) {
			continue
		}
		links = append(links, u.Path)
	}
	return links, nil
}

func (c *Client) getLinksFromCSS(css string) ([]string, error) {
	css = c.chop(css)
	r := regexp.MustCompile(`url\((/[^/].+?)\)`)
	results := r.FindAllStringSubmatch(css, -1)
	var links []string
	for _, result := range results {
		u, _ := c.parseURL(result[1])
		links = append(links, u.Path)
	}
	return links, nil
}

func (c *Client) chop(s string) string {
	var r = regexp.MustCompile(`\r\n|\r|\n`) //throw panic if fail
	return r.ReplaceAllString(s, "")
}
