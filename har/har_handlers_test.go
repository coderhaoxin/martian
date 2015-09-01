// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package har

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
	"github.com/google/martian/session"
)

func TestExportHandlerServeHTTP(t *testing.T) {
	logger := NewLogger("martian", "2.0.0")

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	ctx, err := session.FromContext(nil)
	if err != nil {
		t.Fatalf("session.FromContext(): got %v, want no error", err)
	}
	martian.SetContext(req, ctx)
	defer martian.RemoveContext(req)

	if err := logger.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	res := proxyutil.NewResponse(200, nil, req)
	if err := logger.ModifyResponse(res); err != nil {
		t.Fatalf("ModifyResponse(): got %v, want no error", err)
	}

	h := NewExportHandler(logger)

	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if got, want := rw.Code, http.StatusOK; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}

	hl := &HAR{}
	if err := json.Unmarshal(rw.Body.Bytes(), hl); err != nil {
		t.Fatalf("json.Unmarshal(): got %v, want no error", err)
	}

	if got, want := len(hl.Log.Entries), 1; got != want {
		t.Fatalf("len(hl.Log.Entries): got %v, want %v", got, want)
	}

	entry := hl.Log.Entries[0]
	if got, want := entry.Request.URL, "http://example.com"; got != want {
		t.Errorf("Request.URL: got %q, want %q", got, want)
	}
	if got, want := entry.Response.Status, 200; got != want {
		t.Errorf("Response.Status: got %d, want %d", got, want)
	}

	rh := NewResetHandler(logger)
	req, err = http.NewRequest("DELETE", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	rw = httptest.NewRecorder()
	rh.ServeHTTP(rw, req)

	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest(): got %v, want no error", err)
	}

	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if got, want := rw.Code, http.StatusOK; got != want {
		t.Errorf("rw.Code: got %d, want %d", got, want)
	}

	hl = &HAR{}
	if err := json.Unmarshal(rw.Body.Bytes(), hl); err != nil {
		t.Fatalf("json.Unmarshal(): got %v, want no error", err)
	}

	if got, want := len(hl.Log.Entries), 0; got != want {
		t.Errorf("len(Log.Entries): got %v, want %v", got, want)
	}
}
