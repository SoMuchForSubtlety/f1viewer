package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
)

const (
	proxyMaxRetries = 10
	proxyRetryDelay = 50 * time.Millisecond
	ErrNotRequired  = ProxyError("domain does not use cookies")
)

type ProxyError string

func (e ProxyError) Error() string {
	return string(e)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewProxyServer(streamURL string, logger util.Logger) (*ProxyServer, error) {
	proxy, err := newProxy(streamURL, logger)
	if err != nil {
		return nil, err
	}
	srv := &ProxyServer{
		srv:  http.Server{Handler: proxy},
		log:  logger,
		path: proxy.url.Path,
	}

	return srv, nil
}

type ProxyServer struct {
	srv  http.Server
	log  util.Logger
	path string
}

func (s *ProxyServer) Listen(ctx context.Context) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to start proxy: %w", err)
	}
	go func() {
		err := s.srv.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			s.log.Error(err)
		} else {
			s.log.Info("stopped proxy")
		}
	}()

	go func() {
		<-ctx.Done()
		s.log.Info("closing proxy")
		if err := s.srv.Close(); err != nil {
			s.log.Errorf("failed to stop proxy: %v", err)
		}
	}()
	s.log.Infof("proxy listening at: %s", listener.Addr())

	return "http://" + listener.Addr().String() + s.path, nil
}

func newProxy(streamURL string, logger util.Logger) (*proxy, error) {
	u, err := url.Parse(streamURL)
	if err != nil {
		return nil, err
	}

	j, _ := cookiejar.New(nil)
	c := &http.Client{
		Jar: j,
	}

	res, err := c.Get(streamURL)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if len(j.Cookies(u)) == 0 {
		return nil, ErrNotRequired
	}

	return &proxy{u, b, c, logger}, nil
}

type proxy struct {
	url      *url.URL
	playlist []byte
	client   *http.Client
	util.Logger
}

func (t *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/index.m3u8":
		http.Redirect(w, r, t.url.Path, http.StatusMovedPermanently)
		return
	case t.url.Path:
		_, err := w.Write(t.playlist)
		if err != nil {
			t.Errorf("failed to write http response: %v", err)
		}
		return
	}

	u := *r.URL
	u.Scheme = t.url.Scheme
	u.Host = t.url.Host

	for i := 0; i < proxyMaxRetries; i++ {
		res, err := t.client.Get(u.String())
		if err != nil {
			t.Errorf("upstream proxy request: %v", err)
			time.Sleep(proxyRetryDelay)
			continue
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("received non 200 response code: %d, %s", res.StatusCode, u.String())
			time.Sleep(proxyRetryDelay)
			continue
		}

		b := &bytes.Buffer{}
		if _, err := io.Copy(b, res.Body); err != nil {
			t.Errorf("loading upstream proxy response: %v", err)
			time.Sleep(proxyRetryDelay)
			continue
		}

		if _, err := io.Copy(w, b); err != nil {
			t.Errorf("delivering proxy response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	t.Errorf(`proxy request failed: "%s"`, r.URL)
	http.Error(w, "max retries", http.StatusBadGateway)
}
