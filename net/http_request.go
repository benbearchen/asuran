package net

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
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

func NewHttp(reqUrl string, r *http.Request, dial func(netw, addr string) (net.Conn, error), dont302 bool) (*HttpResponse, []byte, string, error) {
	client := &http.Client{Transport: &http.Transport{Dial: dial}, CheckRedirect: checkRedirect(dont302)}

	var postBody []byte
	var body io.Reader = nil
	if r != nil && r.Method == "POST" {
		b, err := ioutil.ReadAll(r.Body)
		if err == nil {
			postBody = b
			body = bytes.NewReader(b)
		}
	}

	method := "GET"
	if r != nil {
		method = r.Method
	}

	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return nil, postBody, "", err
	}

	if r != nil {
		req.Header = r.Header
	}

	resp, err := client.Do(req)
	if err != nil {
		if urlError, ok := err.(*url.Error); ok {
			if _, ok := urlError.Err.(*redirectError); ok {
				return nil, postBody, urlError.URL, nil
			}
		}

		return nil, postBody, "", err
	}

	return &HttpResponse{resp}, postBody, "", nil
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

func (r *HttpResponse) ProxyReturn(w http.ResponseWriter, wrap io.Writer, recvFirst, forceChunked bool) ([]byte, error) {
	defer r.resp.Body.Close()
	h := w.Header()
	for k, v := range r.Header() {
		h[k] = v
	}

	if forceChunked {
		h.Del("Content-Length")
	}

	if wrap == nil {
		wrap = w
	}

	if recvFirst {
		bytes, err := ioutil.ReadAll(r.resp.Body)
		if err == nil {
			if !forceChunked {
				w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
			}

			w.WriteHeader(r.ResponseCode())
			_, err = wrap.Write(bytes)
		} else {
			w.WriteHeader(502)
		}

		return bytes, err
	} else {
		w.WriteHeader(r.ResponseCode())

		var b bytes.Buffer
		_, err := io.Copy(io.MultiWriter(&b, wrap), r.resp.Body)
		return b.Bytes(), err
	}
}

type redirectError struct {
}

func (e *redirectError) Error() string {
	return "error for mark redirection"
}

func checkRedirect(dont302 bool) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if dont302 {
			return &redirectError{}
		}

		if len(via) > 0 {
			req.Header = via[0].Header
		}

		return nil
	}
}
