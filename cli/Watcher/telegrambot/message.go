package telegrambot

import (
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
)

const (
	SendMessageFormat = "https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s"
)

func SendMessage(botToken, chatID, msg string) error {
	// create client
	proxyUrl := "192.168.3.31:8388"
	dialer, err := proxy.SOCKS5("tcp", proxyUrl, nil, proxy.Direct)
	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}
	transport := &http.Transport{DialContext: dialContext,
		DisableKeepAlives: true}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("POST", fmt.Sprintf(SendMessageFormat, botToken, chatID, msg), nil)
	if err != nil {
		return fmt.Errorf("error in creating http request: %v", err)
	}

	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error in executing http request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error status code: %v", resp.StatusCode)
	}

	return nil
}
