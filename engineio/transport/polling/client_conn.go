package polling

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/utils"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/googollee/go-socket.io/engineio/payload"

	"github.com/googollee/go-socket.io/engineio/transport"
)

type serverConn struct {
	//transport transport.Transporter
	httpClient   *http.Client
	request      *http.Request

	*payload.Payload
	supportBinary bool

	remoteHeader http.Header
	localAddr    net.Addr
	remoteAddr   net.Addr
	url          url.URL
	jsonp        string
}

func newConn(r *http.Request) transport.Conn {
	query := r.URL.Query()
	jsonp := query.Get("j")

	supportBinary := query.Get("b64") == "" && jsonp == ""

	return &serverConn{
		Payload:       payload.New(supportBinary),
		//transport:     transport,
		supportBinary: supportBinary,
		remoteHeader:  r.Header,
		localAddr:     Addr{r.Host},
		remoteAddr:    Addr{r.RemoteAddr},
		url:           *r.URL,
		jsonp:         jsonp,
	}
}

func (c *serverConn) URL() url.URL {
	return c.url
}

func (c *serverConn) SetHeaders(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.UserAgent(), ";MSIE") || strings.Contains(r.UserAgent(), "Trident/") {
		w.Header().Set("X-XSS-Protection", "0")
	}

	//just in case the default behaviour gets changed and it has to handle an origin check
	checkOrigin := Default.CheckOrigin
	if c.transport.CheckOrigin != nil {
		checkOrigin = c.transport.CheckOrigin
	}

	if checkOrigin != nil && checkOrigin(r) {
		if r.URL.Query().Get("j") == "" {
			origin := r.Header.Get("Origin")
			if origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
	}
}

func (c *serverConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		if r.URL.Query().Get("j") == "" {
			c.SetHeaders(w, r)

			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			w.WriteHeader(http.StatusOK)
		}
	case http.MethodGet:
		c.SetHeaders(w, r)

		if jsonp := r.URL.Query().Get("j"); jsonp != "" {
			buf := bytes.NewBuffer(nil)

			if err := c.Payload.FlushOut(buf); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")

			//todo: usage bytes.Buffer ?
			_, _ = w.Write([]byte("___eio[" + jsonp + "](\""))
			_, _ = w.Write([]byte(template.JSEscapeString(buf.String())))
			_, _ = w.Write([]byte("\");"))

			return
		}

		headerVal := "text/plain; charset=UTF-8"
		if c.supportBinary {
			headerVal = "application/octet-stream"
		}

		w.Header().Set("Content-Type", headerVal)


		if err := c.Payload.FlushOut(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case http.MethodPost:
		c.SetHeaders(w, r)

		mime := r.Header.Get("Content-Type")
		supportBinary, err := mimeSupportBinary(mime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.Payload.FeedIn(r.Body, supportBinary); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = w.Write([]byte("ok"))
	default:
		http.Error(w, "invalid method", http.StatusBadRequest)
	}
}

type clientConn struct {
	*payload.Payload

	httpClient   *http.Client
	request      http.Request
	remoteHeader atomic.Value
}

func dial(client *http.Client, url url.URL, header http.Header) (*clientConn, error) {
	if client == nil {
		client = &http.Client{}
	}

	req, err := http.NewRequest("", url.String(), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}

	supportBinary := req.URL.Query().Get("b64") == ""

	if supportBinary {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	}

	return &clientConn{
		Payload:    payload.New(supportBinary),
		httpClient: client,
		request:    *req,
	}, nil
}

func (c *clientConn) Open() (transport.ConnParams, error) {
	go c.getOpen()

	_, pt, r, err := c.NextReader()
	if err != nil {
		return transport.ConnParams{}, err
	}
	if pt != packet.OPEN {
		r.Close()

		return transport.ConnParams{}, errors.New("invalid open")
	}

	conn, err := transport.ReadConnParameters(r)

	if err != nil {
		r.Close()
		return transport.ConnParams{}, err
	}

	err = r.Close()
	if err != nil {
		return transport.ConnParams{}, err
	}

	query := c.request.URL.Query()
	query.Set("sid", conn.SID)
	c.request.URL.RawQuery = query.Encode()

	go c.serveGet()
	go c.servePost()

	return conn, nil
}

func (c *clientConn) URL() url.URL {
	return *c.request.URL
}

func (c *clientConn) LocalAddr() net.Addr {
	return Addr{""}
}

func (c *clientConn) RemoteAddr() net.Addr {
	return Addr{c.request.Host}
}

func (c *clientConn) RemoteHeader() http.Header {
	ret := c.remoteHeader.Load()
	if ret == nil {
		return nil
	}
	return ret.(http.Header)
}

func (c *clientConn) Resume() {
	c.Payload.Resume()

	go c.serveGet()
	go c.servePost()
}

func (c *clientConn) servePost() {
	var buf bytes.Buffer

	req := c.request
	url := *req.URL

	req.URL = &url
	query := url.Query()

	req.Method = "POST"
	req.Body = ioutil.NopCloser(&buf)

	for {
		buf.Reset()
		if err := c.Payload.FlushOut(&buf); err != nil {
			return
		}
		query.Set("t", utils.Timestamp())
		req.URL.RawQuery = query.Encode()
		resp, err := c.httpClient.Do(&req)
		if err != nil {
			c.Payload.Store("post", err)
			c.Close()
			return
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.Payload.Store("post", fmt.Errorf("invalid response: %s(%d)", resp.Status, resp.StatusCode))
			c.Close()
			return
		}
		c.remoteHeader.Store(resp.Header)
	}
}

func (c *clientConn) getOpen() {
	req := c.request
	query := req.URL.Query()
	url := *req.URL

	req.URL = &url
	req.Method = "GET"

	query.Set("t", utils.Timestamp())
	req.URL.RawQuery = query.Encode()
	resp, err := c.httpClient.Do(&req)
	if err != nil {
		c.Payload.Store("get", err)
		c.Close()
		return
	}

	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid request: %s(%d)", resp.Status, resp.StatusCode)
	}

	var supportBinary bool
	if err == nil {
		mime := resp.Header.Get("Content-Type")
		supportBinary, err = mimeSupportBinary(mime)
	}

	if err != nil {
		c.Payload.Store("get", err)
		c.Close()
		return
	}

	c.remoteHeader.Store(resp.Header)

	if err = c.Payload.FeedIn(resp.Body, supportBinary); err != nil {
		return
	}
}

func (c *clientConn) serveGet() {
	req := c.request
	query := req.URL.Query()

	url := *req.URL
	req.URL = &url

	req.Method = http.MethodGet

	for {
		query.Set("t", utils.Timestamp())
		req.URL.RawQuery = query.Encode()
		resp, err := c.httpClient.Do(&req)
		if err != nil {
			c.Payload.Store("get", err)
			c.Close()
			return
		}

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("invalid request: %s(%d)", resp.Status, resp.StatusCode)
		}

		var supportBinary bool
		if err == nil {
			mime := resp.Header.Get("Content-Type")
			supportBinary, err = mimeSupportBinary(mime)
		}

		if err != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			c.Payload.Store("get", err)
			c.Close()
			return
		}

		if err = c.Payload.FeedIn(resp.Body, supportBinary); err != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			return
		}

		c.remoteHeader.Store(resp.Header)
	}
}

func (c *serverConn) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (c *serverConn) Write(b []byte) (n int, err error) {
	panic("implement me")
}

func (c *serverConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (c *serverConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (c *serverConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (c *serverConn) NextRead() (frame.Type, packet.Type, io.ReadCloser, error) {
	panic("implement me")
}

func (c *serverConn) NextWrite(f frame.Type, p packet.Type) (io.WriteCloser, error) {
	panic("implement me")
}