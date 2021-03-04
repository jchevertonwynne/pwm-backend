package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_registerHandler(t *testing.T) {
	type args struct {
		method        string
		existingUsers map[string]string
		username      string
		password      string
	}
	type expected struct {
		code int
		body string
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			"can create user",
			args{
				http.MethodPost,
				make(map[string]string),
				"joseph",
				"password123",
			},
			expected{
				http.StatusOK,
				"",
			},
		},
		{
			"can't create user as user already exists",
			args{
				http.MethodPost,
				map[string]string{"joseph": "hello"},
				"joseph",
				"password123",
			},
			expected{
				http.StatusBadRequest,
				"Username is taken\n",
			},
		},
		{
			"can't create user as only POST allowed",
			args{
				http.MethodGet,
				make(map[string]string),
				"joseph",
				"password123",
			},
			expected{
				http.StatusMethodNotAllowed,
				"Request is not a POST\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users = tt.args.existingUsers

			user := registerRequest{
				Username: tt.args.username,
				Password: tt.args.password,
			}
			jsonified, err := json.Marshal(user)
			if err != nil {
				t.Fatal(err)
			}

			r, err := http.NewRequest(tt.args.method, "/", bytes.NewReader(jsonified))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			registerHandler(w, r)
			res := w.Result()

			if res.StatusCode != tt.expected.code {
				t.Errorf("expected response code %d, got %d", tt.expected.code, res.StatusCode)
			}

			bodyContents, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(bodyContents) != tt.expected.body {
				t.Errorf("exptected body %q, got %q", string(bodyContents), tt.expected.body)
			}
		})
	}
}
