/*
Copyright 2014-2016 Pinterest, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package http

import (
	"net/http"

	"github.com/pinterest/bender"
)

// ResponseValidator validates an HTTP response.
type ResponseValidator func(request interface{}, resp *http.Response) error

// CreateExecutor creates an HTTP request executor.
func CreateExecutor(tr *http.Transport, client *http.Client, responseValidator ResponseValidator) bender.RequestExecutor {
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
		err = responseValidator(request, resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}
