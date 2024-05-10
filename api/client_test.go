package api

import (
	"net/http"

	"github.com/jialequ/mplb/pkg/httpmock"
)

func newTestClient(reg *httpmock.Registry) *Client {
	client := &http.Client{}
	httpmock.ReplaceTripper(client, reg)
	return NewClientFromHTTP(client)
}
