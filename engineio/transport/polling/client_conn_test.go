package polling

import (
"bytes"
"fmt"
	"github.com/googollee/go-socket.io/engineio/transport"
	"io/ioutil"
"net/http"
"net/http/httptest"
"net/url"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

"github.com/googollee/go-socket.io/engineio/frame"
"github.com/googollee/go-socket.io/engineio/packet"
)

func TestDialOpen(t *testing.T) {
	cp := transport.ConnParams{
		PingInterval: time.Second,
		PingTimeout:  time.Minute,
		SID:          "abcdefg",
		Upgrades:     []string{"polling"},
	}

	should := assert.New(t)
	must := require.New(t)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		query := r.URL.Query()
		should.NotEmpty(r.URL.Query().Get("t"))
		sid := query.Get("sid")

		if sid == "" {
			buf := bytes.NewBuffer(nil)
			cp.WriteTo(buf)

			fmt.Fprintf(w, "%d", buf.Len()+1)

			w.Write([]byte(":0"))
			w.Write(buf.Bytes())

			return
		}

		if r.Method == http.MethodPost {
			must.Equal(cp.SID, sid)

			b, err := ioutil.ReadAll(r.Body)
			must.Nil(err)
			should.Equal("6:4hello", string(b))
		}
	}

	httpSvr := httptest.NewServer(http.HandlerFunc(handler))

	defer httpSvr.Close()

	u, err := url.Parse(httpSvr.URL)
	must.Nil(err)

	query := u.Query()
	query.Set("b64", "1")
	u.RawQuery = query.Encode()

	cc, err := dial(nil, u, nil)
	must.Nil(err)

	defer cc.Close()

	params, err := cc.Open()
	must.Nil(err)
	should.Equal(cp, params)

	ccURL := cc.URL()
	sid := ccURL.Query().Get("sid")
	should.Equal(cp.SID, sid)

	w, err := cc.NextWriter(frame.String, packet.MESSAGE)
	should.Nil(err)

	_, err = w.Write([]byte("hello"))
	should.Nil(err)

	err = w.Close()
	should.Nil(err)
}

func TestServerJSONP(t *testing.T) {
	var scValue atomic.Value

	transport := Default
	conn := make(chan transport.Conn, 1)

	handler := func(w http.ResponseWriter, r *http.Request) {
		c := scValue.Load()
		if c == nil {
			co, err := transport.Accept(w, r)
			require.NoError(t, err)

			scValue.Store(co)
			c = co
			conn <- co
		}
		c.(http.Handler).ServeHTTP(w, r)
	}

	httpSvr := httptest.NewServer(http.HandlerFunc(handler))
	defer httpSvr.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		sc := <-conn

		defer sc.Close()

		w, err := sc.NextWriter(frame.Binary, packet.MESSAGE)
		require.NoError(t, err)

		_, err = w.Write([]byte("hello"))
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)

		w, err = sc.NextWriter(frame.String, packet.MESSAGE)
		require.NoError(t, err)

		_, err = w.Write([]byte("world"))
		require.NoError(t, err)

		err = w.Close()
		require.NoError(t, err)
	}()

	{
		u := httpSvr.URL + "?j=jsonp_f1"
		resp, err := http.Get(u)
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, "text/javascript; charset=UTF-8", resp.Header.Get("Content-Type"))
		bs, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("___eio[jsonp_f1](\"%s\");", template.JSEscapeString("10:b4aGVsbG8=")), string(bs))
	}
	{
		u := httpSvr.URL + "?j=jsonp_f2"
		resp, err := http.Get(u)
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, "text/javascript; charset=UTF-8", resp.Header.Get("Content-Type"))

		bs, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "___eio[jsonp_f2](\"6:4world\");", string(bs))
	}

	wg.Wait()
}
