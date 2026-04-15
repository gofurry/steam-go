package transport

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestWrapRoundTripperSupportsHTTPSProxy(t *testing.T) {
	t.Parallel()

	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ok`))
	}))
	defer target.Close()

	var proxyCalls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			t.Fatalf("unexpected proxy method: %s", r.Method)
		}
		proxyCalls.Add(1)

		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("proxy response writer does not support hijacking")
		}
		clientConn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Hijack returned error: %v", err)
		}

		targetConn, err := net.Dial("tcp", target.Listener.Addr().String())
		if err != nil {
			_ = clientConn.Close()
			t.Fatalf("Dial target returned error: %v", err)
		}

		_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		go proxyCopy(clientConn, targetConn)
		go proxyCopy(targetConn, clientConn)
	}))
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatalf("Parse proxy URL returned error: %v", err)
	}

	base := target.Client().Transport.(*http.Transport)
	client := &http.Client{
		Transport: WrapRoundTripper(base, staticProxySelector{url: proxyURL}),
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do returned error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if got, want := string(body), "ok"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if proxyCalls.Load() == 0 {
		t.Fatal("expected proxy CONNECT to be used")
	}
}

func proxyCopy(dst net.Conn, src net.Conn) {
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	_ = src.Close()
}

type staticProxySelector struct {
	url *url.URL
}

func (s staticProxySelector) Next(*http.Request) (*url.URL, error) {
	return s.url, nil
}

func TestCloneTransportCopiesConfig(t *testing.T) {
	t.Parallel()

	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			ServerName: "example.com",
		},
		ProxyConnectHeader: http.Header{
			"X-Test": []string{"value"},
		},
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{
			"h2": func(string, *tls.Conn) http.RoundTripper { return nil },
		},
	}

	cloned := cloneTransport(base)
	if cloned == base {
		t.Fatal("expected a distinct transport instance")
	}
	if cloned.TLSClientConfig == base.TLSClientConfig {
		t.Fatal("expected TLSClientConfig to be cloned")
	}
	if cloned.ProxyConnectHeader == nil || cloned.ProxyConnectHeader.Get("X-Test") != "value" {
		t.Fatalf("unexpected ProxyConnectHeader: %#v", cloned.ProxyConnectHeader)
	}
	cloned.ProxyConnectHeader.Set("X-Test", "changed")
	if got, want := base.ProxyConnectHeader.Get("X-Test"), "value"; got != want {
		t.Fatalf("base ProxyConnectHeader mutated: got %q want %q", got, want)
	}
	if cloned.TLSNextProto == nil || len(cloned.TLSNextProto) != 1 {
		t.Fatalf("unexpected TLSNextProto: %#v", cloned.TLSNextProto)
	}
	if &cloned.TLSNextProto == &base.TLSNextProto {
		t.Fatal("expected TLSNextProto map to be copied")
	}
}
