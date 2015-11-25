/*
Copyright 2014 Pinterest.com

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

	"github.com/benbooth493/bender"
)

type HttpValidator func(request interface{}, resp *http.Response, target interface{}) error

func CreateHttpExecutor(tr *http.Transport, client *http.Client, httpValidator HttpValidator, target interface{}) bender.RequestExecutor {
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
		err = httpValidator(request, resp, target)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}
