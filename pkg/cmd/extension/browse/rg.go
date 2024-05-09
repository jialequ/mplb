package browse

import (
	"net/http"
	"time"

	"github.com/jialequ/mplb/api"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/repo/view"
)

type readmeGetter struct {
	client *http.Client
}

func newReadmeGetter(client *http.Client, cacheTTL time.Duration) *readmeGetter {
	cachingClient := api.NewCachedHTTPClient(client, cacheTTL)
	return &readmeGetter{
		client: cachingClient,
	}
}

func (g *readmeGetter) Get(repoFullName string) (string, error) {
	repo, err := ghrepo.FromFullName(repoFullName)
	if err != nil {
		return "", err
	}
	readme, err := view.RepositoryReadme(g.client, repo, "")
	if err != nil {
		return "", err
	}
	return readme.Content, nil
}
