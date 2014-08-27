package http

import (
	"net/http"
	"io"
	"github.com/Pinterest/bender"
)

type HttpBodyHandler func(*io.ReadCloser) error

func CreateHttpExecutor(tr *http.Transport, client *http.Client, bodyHandler HttpBodyHandler) bender.RequestExecutor {
	if tr == nil {
		tr = &http.Transport{}
		client = &http.Client{Transport: tr}
	} else if client == nil {
		client = &http.Client{Transport: tr}
	}

	return func(_ int64, request interface{}) error {
		req := request.(*http.Request)
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
