package proxy_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/jasonlvhit/gocron"
	"github.com/motemen/go-loghttp"
)

// Client is used to implement http proxy client
type Client struct {
	*http.Client
	cfg     Config
	logger  logger
	mu      sync.Mutex
	proxies []proxy
}

type logger interface {
	Error(args ...interface{})
}

type proxy struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Type int    `json:"type"`
}

// Options provides request options
type Options struct {
	Headers map[string]string
	Params  map[string]string
	Data    []byte
}

// NewClient provides new proxy client
func NewClient(cfg Config, logger logger) *Client {

	if logger == nil {
		logger = NewLogger()
	}

	populateConfig(&cfg)

	client := Client{
		Client: &http.Client{Timeout: time.Duration(cfg.ClientTimeOut) * time.Second},
		mu:     sync.Mutex{},
		cfg:    cfg,
		logger: logger,
	}

	if err := client.runRefresher(); err != nil {
		client.logger.Error(err)
		os.Exit(1)
	}

	return &client
}

func populateConfig(cfg *Config) {
	if cfg.ProxyURL == "" {
		cfg.ProxyURL = "http://free-proxy-list.appspot.com/proxy.json"
	}

	if cfg.MaxConn == 0 {
		cfg.MaxConn = 1024
	}

	if cfg.ClientTimeOut == 0 {
		cfg.ClientTimeOut = 30
	}
}

func (c *Client) runRefresher() error {
	c.refreshProxies()

	err := gocron.Every(1).Hour().Do(func() {
		c.refreshProxies()
	})
	if err != nil {
		return err
	}

	go func() {
		<-gocron.Start()
	}()

	return nil
}

func (c *Client) refreshProxies() {
	res, err := c.DoRequest(c.cfg.ProxyURL, "GET", Options{})
	if err != nil {
		c.logger.Error(err)
	}
	var proxies []proxy
	if err := json.Unmarshal(res, &proxies); err != nil {
		c.logger.Error(err)
	}

	c.mu.Lock()
	c.proxies = proxies
	c.mu.Unlock()
}

// DoRequest - do request
func (c *Client) DoRequest(requestURL, method string, opts Options) ([]byte, error) {

	req, err := c.request(requestURL, method, opts)
	if err != nil {
		return nil, err
	}

	proxyURL, err := c.proxyURL()
	if err != nil {
		c.logger.Error(err)
	}

	tr := loghttp.Transport{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: c.cfg.MaxConn,
			TLSHandshakeTimeout: time.Duration(c.cfg.HandshakeTimeout),
			Proxy:               http.ProxyURL(proxyURL),
		},
		LogRequest:  c.cfg.LogRequest,
		LogResponse: c.cfg.LogResponse,
	}

	c.Transport = &tr

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("response body is empty")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d; url %s; response error body: %s", resp.StatusCode, requestURL, body)
	}

	return body, nil
}

func (c *Client) proxyURL() (*url.URL, error) {
	if len(c.proxies) == 0 {
		return nil, fmt.Errorf("empty proxy list")
	}

	var u string
	attempt := 0
	c.mu.Lock()
	for {
		if attempt >= 30 {
			break
		}
		attempt++
		rand.Seed(time.Now().Unix())
		n := rand.Int() % len(c.proxies)
		if c.proxies[n].Type == 0 {
			u = "http://" + c.proxies[n].Host + ":" + strconv.Itoa(c.proxies[n].Port)
			break
		}
	}
	c.mu.Unlock()

	proxyURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return proxyURL, nil
}

func (c *Client) request(url, method string, opts Options) (*http.Request, error) {
	if url == "" {
		return nil, fmt.Errorf("%s", "requst URL is empty")
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(opts.Data))
	if err != nil {
		return nil, err
	}
	if req.URL == nil {
		return nil, fmt.Errorf("empty request URL object")
	}
	q := req.URL.Query()
	for k, v := range opts.Params {
		q.Add(k, v)
	}
	if q != nil {
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range opts.Headers {
		req.Header.Add(k, v)
	}
	req.Header.Add("X-Forwarded-For", randomdata.IpV4Address())

	return req, nil
}

// Stop stops cron job
func (c *Client) Stop() {
	gocron.Clear()
}
