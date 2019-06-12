/*
 * Copyright 2019 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package schema

import (
	"encoding/json"
	"io"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/gqlerror"
)

// GraphQL spec on errors is here https://graphql.github.io/graphql-spec/June2018/#sec-Errors

// Response represents a GraphQL response
type Response struct {
	// TODO: dgraph response type (x.go) is similar, should I be leaning on that?
	// ATM, no ,cause I'm trying to follow the spec really closely with errors and
	// when data should be missing or "null"
	Errors gqlerror.List   `json:"errors,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// ErrorResponsef returns a Response containing a single GraphQL error with a message
// obtained by Sprintf-ing the argugments
func ErrorResponsef(format string, args ...interface{}) *Response {
	return &Response{
		Errors: gqlerror.List{gqlerror.Errorf(format, args...)},
	}
}

// WithNullData sets the data response of r such that subsequent calls
// to r.WriteTo will write `"data": null`
func (r *Response) WithNullData() {
	r.Data = json.RawMessage(`null`)
}

// WriteTo writes the GraphQL response as unindented JSON to w
// and returns the number of bytes written and error, if any.
func (r *Response) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(r)
	if err != nil {
		// probably indicatesa bug that's written invalid bytes to r.Data
		// should I even do it this way - why not just write the bytes directly to w?
		msg := "Failed to write a valid GraphQL JSON response"
		glog.Errorf(msg, err) // also dump in other debugging like r into V(2)?
		errResp := ErrorResponsef(msg)
		errResp.WithNullData()
		b, err = json.Marshal(errResp)
		if err != nil {
			return 0, errors.Wrap(err, "failed to even marshal json error msg")
		}
	}

	n, err := w.Write(b)
	return int64(n), err
}
