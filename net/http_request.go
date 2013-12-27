package net

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type HttpResponse struct {
	resp *http.Response
}

func NewHttpGet(url string) (*HttpResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{resp}, nil
}

func NewHttp(url string, r *http.Request) (*HttpResponse, error) {
	client := &http.Client{}

	var body io.Reader = nil
	if r.Method == "POST" {
		b, err := ioutil.ReadAll(r.Body)
		if err == nil {
			body = bytes.NewReader(b)
		}
	}

	req, err := http.NewRequest(r.Method, url, body)
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{resp}, nil
}

func (r *HttpResponse) Close() {
	if r.resp != nil {
		r.resp.Body.Close()
	}
}

func (r *HttpResponse) ReadAllBytes() ([]byte, error) {
	if r.resp != nil {
		return ioutil.ReadAll(r.resp.Body)
	}

	return make([]byte, 0), fmt.Errorf("invalid response")
}

func (r *HttpResponse) ReadAll() (string, error) {
	bytes, err := r.ReadAllBytes()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (r *HttpResponse) Header() http.Header {
	return r.resp.Header
}

func (r *HttpResponse) ResponseCode() int {
	return r.resp.StatusCode
}
