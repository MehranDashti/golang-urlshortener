package testhelper

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// MakeRequest fires an HTTP request against the test router
// and returns the recorder so you can inspect status + body
func MakeRequest(r *gin.Engine, method, path, body string, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	} else {
		reqBody = bytes.NewBufferString("")
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ParseBody unmarshals response body into a map
func ParseBody(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		panic("ParseBody: failed to unmarshal response body: " + err.Error())
	}
	return result
}

// GetData extracts the "data" field from a standard API response
func GetData(w *httptest.ResponseRecorder) map[string]interface{} {
	body := ParseBody(w)
	if data, ok := body["data"].(map[string]interface{}); ok {
		return data
	}
	return nil
}
