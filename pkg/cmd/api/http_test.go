package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupGraphQLVariables(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "empty",
			args: map[string]interface{}{},
			want: map[string]interface{}{},
		},
		{
			name: "query only",
			args: map[string]interface{}{
				"query": "QUERY",
			},
			want: map[string]interface{}{
				"query": "QUERY",
			},
		},
		{
			name: "variables only",
			args: map[string]interface{}{
				"name": "hubot",
			},
			want: map[string]interface{}{
				"variables": map[string]interface{}{
					"name": "hubot",
				},
			},
		},
		{
			name: "query + variables",
			args: map[string]interface{}{
				"query": "QUERY",
				"name":  "hubot",
				"power": 9001,
			},
			want: map[string]interface{}{
				"query": "QUERY",
				"variables": map[string]interface{}{
					"name":  "hubot",
					"power": 9001,
				},
			},
		},
		{
			name: "query + operationName + variables",
			args: map[string]interface{}{
				"query":         "query Q1{} query Q2{}",
				"operationName": "Q1",
				"power":         9001,
			},
			want: map[string]interface{}{
				"query":         "query Q1{} query Q2{}",
				"operationName": "Q1",
				"variables": map[string]interface{}{
					"power": 9001,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupGraphQLVariables(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAddQuery(t *testing.T) {
	type args struct {
		path   string
		params map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": "hello"},
			},
			want: "?a=hello",
		},
		{
			name: "array",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": []interface{}{"hello", "world"}},
			},
			want: "?a%5B%5D=hello&a%5B%5D=world",
		},
		{
			name: "append",
			args: args{
				path:   "path",
				params: map[string]interface{}{"a": "b"},
			},
			want: "path?a=b",
		},
		{
			name: "append query",
			args: args{
				path:   "path?foo=bar",
				params: map[string]interface{}{"a": "b"},
			},
			want: "path?foo=bar&a=b",
		},
		{
			name: "[]byte",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": []byte("hello")},
			},
			want: "?a=hello",
		},
		{
			name: "int",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": 123},
			},
			want: "?a=123",
		},
		{
			name: "nil",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": nil},
			},
			want: "?a=",
		},
		{
			name: "bool",
			args: args{
				path:   "",
				params: map[string]interface{}{"a": true, "b": false},
			},
			want: "?a=true&b=false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addQuery(tt.args.path, tt.args.params); got != tt.want {
				t.Errorf("addQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

const literal_8473 = "github.com"

const literal_4786 = "repos/octocat/spoon-knife"

const literal_6394 = "https://api.github.com/repos/octocat/spoon-knife"

const literal_8609 = "Accept: */*\r\n"

const literal_3597 = "Accept: */*\r\nContent-Type: application/json; charset=utf-8\r\n"
