package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

var HTTPClient = &http.Client{
	Transport: &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        200,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func CheckUrl(urlStr string, ctx context.Context) (string, error) {
	fullURL := fmt.Sprintf("https://%s", urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fullURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		if urlStr == "vk.com" || urlStr == "wikipedia.org" {
			fmt.Printf("Специальная обработка ошибки для %s: %v\n", urlStr, err)
		}
		if ctx.Err() == context.DeadlineExceeded {
			//Сомнения, что нужно по смерти контекста возвращать не доступен, ведь контекст мог умереть по таймауту?
			fmt.Println("Контекст истек для URL:", urlStr)
			return "not available", nil
		}
		var ue *url.Error
		if errors.As(err, &ue) {
			return "not available", nil
		}
		var netErr *net.Error
		if errors.As(err, &netErr) {
			return "not available", nil
		}
		return "", err
	}
	defer resp.Body.Close()
	isAvailable := resp.StatusCode >= 200 && resp.StatusCode < 400
	if urlStr == "vk.com" || urlStr == "wikipedia.org" {
		fmt.Printf("Специальная обработка статуса для %s: %v\n", urlStr, resp.StatusCode)
	}
	if isAvailable {
		return "available", nil
	} else {
		return "not available", nil
	}
}
