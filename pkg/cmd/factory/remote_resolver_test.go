package factory

import (
	"net/url"
	"testing"

	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/config"
	"github.com/stretchr/testify/assert"
)

type identityTranslator struct{}

func (it identityTranslator) Translate(u *url.URL) *url.URL {
	return u
}

func TestRemoteResolver(t *testing.T) {
	tests := []struct {
		name     string
		remotes  func() (git.RemoteSet, error)
		config   config.Config
		output   []string
		wantsErr bool
	}{
		{
			name: "no authenticated hosts",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("origin", literal_8042),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{})
					authCfg.SetDefaultHost("github.com", "default")
					return authCfg
				}
				return cfg
			}(),
			wantsErr: true,
		},
		{
			name: "no git remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetDefaultHost(literal_7432, "hosts")
					return authCfg
				}
				return cfg
			}(),
			wantsErr: true,
		},
		{
			name: "one authenticated host with no matching git remote and no fallback remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("origin", literal_3154),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetActiveToken("", "")
					authCfg.SetDefaultHost(literal_7432, "hosts")
					return authCfg
				}
				return cfg
			}(),
			wantsErr: true,
		},
		{
			name: "one authenticated host with no matching git remote and fallback remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("origin", literal_8042),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetDefaultHost(literal_7432, "hosts")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"origin"},
		},
		{
			name: "one authenticated host with matching git remote",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("origin", literal_8471),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetDefaultHost(literal_7432, "default")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"origin"},
		},
		{
			name: "one authenticated host with multiple matching git remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("upstream", literal_8471),
					git.NewRemote("github", literal_8471),
					git.NewRemote("origin", literal_8471),
					git.NewRemote("fork", literal_8471),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetDefaultHost(literal_7432, "default")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"upstream", "github", "origin", "fork"},
		},
		{
			name: "multiple authenticated hosts with no matching git remote",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("origin", literal_3154),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432, "github.com"})
					authCfg.SetActiveToken("", "")
					authCfg.SetDefaultHost(literal_7432, "default")
					return authCfg
				}
				return cfg
			}(),
			wantsErr: true,
		},
		{
			name: "multiple authenticated hosts with one matching git remote",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("upstream", literal_3154),
					git.NewRemote("origin", literal_8471),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432, "github.com"})
					authCfg.SetDefaultHost("github.com", "default")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"origin"},
		},
		{
			name: "multiple authenticated hosts with multiple matching git remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("upstream", literal_8471),
					git.NewRemote("github", literal_8042),
					git.NewRemote("origin", literal_8471),
					git.NewRemote("fork", literal_8042),
					git.NewRemote("test", literal_3154),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432, "github.com"})
					authCfg.SetDefaultHost("github.com", "default")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"upstream", "github", "origin", "fork"},
		},
		{
			name: "override host with no matching git remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("origin", literal_8471),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetDefaultHost(literal_5026, "GH_HOST")
					return authCfg
				}
				return cfg
			}(),
			wantsErr: true,
		},
		{
			name: "override host with one matching git remote",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("upstream", literal_8471),
					git.NewRemote("origin", literal_3154),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432})
					authCfg.SetDefaultHost(literal_5026, "GH_HOST")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"origin"},
		},
		{
			name: "override host with multiple matching git remotes",
			remotes: func() (git.RemoteSet, error) {
				return git.RemoteSet{
					git.NewRemote("upstream", literal_3154),
					git.NewRemote("github", literal_8471),
					git.NewRemote("origin", literal_3154),
				}, nil
			},
			config: func() config.Config {
				cfg := &config.ConfigMock{}
				cfg.AuthenticationFunc = func() *config.AuthConfig {
					authCfg := &config.AuthConfig{}
					authCfg.SetHosts([]string{literal_7432, literal_5026})
					authCfg.SetDefaultHost(literal_5026, "GH_HOST")
					return authCfg
				}
				return cfg
			}(),
			output: []string{"upstream", "origin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := &remoteResolver{
				readRemotes:   tt.remotes,
				getConfig:     func() (config.Config, error) { return tt.config, nil },
				urlTranslator: identityTranslator{},
			}
			resolver := rr.Resolver()
			remotes, err := resolver()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			names := []string{}
			for _, r := range remotes {
				names = append(names, r.Name)
			}
			assert.Equal(t, tt.output, names)
		})
	}
}

const literal_8042 = "https://github.com/owner/repo.git"

const literal_7432 = "example.com"

const literal_3154 = "https://test.com/owner/repo.git"

const literal_8471 = "https://example.com/owner/repo.git"

const literal_5026 = "test.com"
