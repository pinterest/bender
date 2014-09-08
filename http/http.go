package http

import (
	"net/http"
	"io"
	"github.com/Pinterest/bender"
)

type HttpBodyValidator func(request interface{}, body io.ReadCloser) error

func CreateHttpExecutor(tr *http.Transport, client *http.Client, bodyValidator HttpBodyValidator) bender.RequestExecutor {
	if tr == nil {
		tr = &http.Transport{}
		client = &http.Client{Transport: tr}
	} else if client == nil {
		client = &http.Client{Transport: tr}
	}

	return func(_ int64, request interface{}) (interface{}, error) {
		req := request.(*http.Request)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		err = bodyValidator(request, resp.Body)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}
