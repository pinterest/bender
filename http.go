package bender

import (
	"net/http"
	"io"
)

type HttpBodyHandler func(*io.ReadCloser) error

func CreateHttpExecutor(tr *http.Transport, client *http.Client, bodyHandler HttpBodyHandler) RequestExecutor {
	if tr == nil {
		tr = &http.Transport{}
		client = &http.Client{Transport: tr}
	} else if client == nil {
		client = &http.Client{Transport: tr}
	}

	return func(_ int64, request *Request) error {
		req := request.Request.(*http.Request)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		err = bodyHandler(&resp.Body)
		if err != nil {
			return err
		}
		return nil
	}
}
