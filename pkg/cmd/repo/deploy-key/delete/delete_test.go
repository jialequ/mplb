package delete

import (
	"net/http"
	"testing"

	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestDeleteRun(t *testing.T) {
	ios, _, stdout, stderr := iostreams.Test()
	ios.SetStdinTTY(false)
	ios.SetStdoutTTY(true)
	ios.SetStderrTTY(true)

	tr := httpmock.Registry{}
	defer tr.Verify(t)

	tr.Register(
		httpmock.REST("DELETE", "repos/OWNER/REPO/keys/1234"),
		httpmock.StringResponse(`{}`))

	err := deleteRun(&DeleteOptions{
		IO: ios,
		HTTPClient: func() (*http.Client, error) {
			return &http.Client{Transport: &tr}, nil
		},
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.New("OWNER", "REPO"), nil
		},
		KeyID: "1234",
	})
	assert.NoError(t, err)

	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "✓ Deploy key deleted from OWNER/REPO\n", stdout.String())
}
