package telegram

import (
	"context"
	"net"
	"net/url"

	"golang.org/x/net/proxy"
)

var dialer = func(ctx context.Context, network, address string) (net.Conn, error) {
	proxyURL, err := url.Parse("socks5://127.0.0.1:7890")
	if err != nil {
		return nil, err
	}
	proxy, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return proxy.Dial(network, address)
}
