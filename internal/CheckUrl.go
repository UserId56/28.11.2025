package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

func CheckUrl(urlStr string, ctx context.Context) (string, error) {
	address := net.JoinHostPort(urlStr, "443")
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	var errTimeOut net.Error
	if err != nil {
		if errors.As(err, &errTimeOut) && errTimeOut.Timeout() {
			return "not available", nil
		}
		fmt.Printf("Ошибка подключения для %s: %v\n", urlStr, err)
		return "", err
	}
	defer conn.Close()
	return "available", nil

}
