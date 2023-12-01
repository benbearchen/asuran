package net

import (
	"net/http"
)

func EasyTunnel(url string, w http.ResponseWriter, r *http.Request) error {
	resp, _, _, err := NewHttp(url, r, nil, false)
	if err != nil {
		return err
	}

	defer resp.Close()
	resp.ProxyReturn(w, nil, false, false)
	return nil
}
