package set

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmd/secret/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdSet(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wants    SetOptions
		stdinTTY bool
		wantsErr bool
	}{
		{
			name:     "invalid visibility",
			cli:      "cool_secret --org coolOrg -v'mistyVeil'",
			wantsErr: true,
		},
		{
			name:     "invalid visibility",
			cli:      "cool_secret --org coolOrg -v'selected'",
			wantsErr: true,
		},
		{
			name:     "repos with wrong vis",
			cli:      "cool_secret --org coolOrg -v'private' -rcoolRepo",
			wantsErr: true,
		},
		{
			name:     "no name",
			cli:      "",
			wantsErr: true,
		},
		{
			name:     "multiple names",
			cli:      "cool_secret good_secret",
			wantsErr: true,
		},
		{
			name:     "visibility without org",
			cli:      "cool_secret -vall",
			wantsErr: true,
		},
		{
			name: "repos without vis",
			cli:  "cool_secret -bs --org coolOrg -rcoolRepo",
			wants: SetOptions{
				SecretName:      "cool_secret",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"coolRepo"},
				Body:            "s",
				OrgName:         "coolOrg",
			},
		},
		{
			name: "org with selected repo",
			cli:  "-ocoolOrg -bs -vselected -rcoolRepo cool_secret",
			wants: SetOptions{
				SecretName:      "cool_secret",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"coolRepo"},
				Body:            "s",
				OrgName:         "coolOrg",
			},
		},
		{
			name: "org with selected repos",
			cli:  `--org=coolOrg -bs -vselected -r="coolRepo,radRepo,goodRepo" cool_secret`,
			wants: SetOptions{
				SecretName:      "cool_secret",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"coolRepo", "goodRepo", "radRepo"},
				Body:            "s",
				OrgName:         "coolOrg",
			},
		},
		{
			name: "user with selected repos",
			cli:  `-u -bs -r"monalisa/coolRepo,cli/cli,github/hub" cool_secret`,
			wants: SetOptions{
				SecretName:      "cool_secret",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"monalisa/coolRepo", literal_5461, "github/hub"},
				Body:            "s",
			},
		},
		{
			name: "repo",
			cli:  `cool_secret -bliteral_6879`,
			wants: SetOptions{
				SecretName: "cool_secret",
				Visibility: shared.Private,
				Body:       literal_6879,
				OrgName:    "",
			},
		},
		{
			name: "env",
			cli:  `cool_secret -bliteral_6879 -eRelease`,
			wants: SetOptions{
				SecretName: "cool_secret",
				Visibility: shared.Private,
				Body:       literal_6879,
				OrgName:    "",
				EnvName:    "Release",
			},
		},
		{
			name: "vis all",
			cli:  `cool_secret --org coolOrg -b"cool" -vall`,
			wants: SetOptions{
				SecretName: "cool_secret",
				Visibility: shared.All,
				Body:       "cool",
				OrgName:    "coolOrg",
			},
		},
		{
			name: "no store",
			cli:  `cool_secret --no-store`,
			wants: SetOptions{
				SecretName: "cool_secret",
				Visibility: shared.Private,
				DoNotStore: true,
			},
		},
		{
			name: "Dependabot repo",
			cli:  `cool_secret -bliteral_6879 --app Dependabot`,
			wants: SetOptions{
				SecretName:  "cool_secret",
				Visibility:  shared.Private,
				Body:        literal_6879,
				OrgName:     "",
				Application: "Dependabot",
			},
		},
		{
			name: "Dependabot org",
			cli:  "-ocoolOrg -bs -vselected -rcoolRepo cool_secret -aDependabot",
			wants: SetOptions{
				SecretName:      "cool_secret",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"coolRepo"},
				Body:            "s",
				OrgName:         "coolOrg",
				Application:     "Dependabot",
			},
		},
		{
			name: "Codespaces org",
			cli:  `random_secret -ocoolOrg -b"random value" -vselected -r"coolRepo,cli/cli" -aCodespaces`,
			wants: SetOptions{
				SecretName:      "random_secret",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"coolRepo", literal_5461},
				Body:            "random value",
				OrgName:         "coolOrg",
				Application:     "Codespaces",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			ios.SetStdinTTY(tt.stdinTTY)

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *SetOptions
			cmd := NewCmdSet(f, func(opts *SetOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.SecretName, gotOpts.SecretName)
			assert.Equal(t, tt.wants.Body, gotOpts.Body)
			assert.Equal(t, tt.wants.Visibility, gotOpts.Visibility)
			assert.Equal(t, tt.wants.OrgName, gotOpts.OrgName)
			assert.Equal(t, tt.wants.EnvName, gotOpts.EnvName)
			assert.Equal(t, tt.wants.DoNotStore, gotOpts.DoNotStore)
			assert.ElementsMatch(t, tt.wants.RepositoryNames, gotOpts.RepositoryNames)
			assert.Equal(t, tt.wants.Application, gotOpts.Application)
		})
	}
}

func TestSetRunrepo(t *testing.T) {
	tests := []struct {
		name    string
		opts    *SetOptions
		wantApp string
	}{
		{
			name: "Actions",
			opts: &SetOptions{
				Application: "actions",
			},
			wantApp: "actions",
		},
		{
			name: "Dependabot",
			opts: &SetOptions{
				Application: "dependabot",
			},
			wantApp: "dependabot",
		},
		{
			name: "defaults to Actions",
			opts: &SetOptions{
				Application: "",
			},
			wantApp: "actions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}

			reg.Register(httpmock.REST("GET", fmt.Sprintf("repos/owner/repo/%s/secrets/public-key", tt.wantApp)),
				httpmock.JSONResponse(PubKey{ID: "123", Key: literal_9163}))

			reg.Register(httpmock.REST("PUT", fmt.Sprintf("repos/owner/repo/%s/secrets/cool_secret", tt.wantApp)),
				httpmock.StatusStringResponse(201, `{}`))

			ios, _, _, _ := iostreams.Test()

			opts := &SetOptions{
				HttpClient: func() (*http.Client, error) {
					return &http.Client{Transport: reg}, nil
				},
				Config: func() (config.Config, error) { return config.NewBlankConfig(), nil },
				BaseRepo: func() (ghrepo.Interface, error) {
					return ghrepo.FromFullName(literal_1049)
				},
				IO:             ios,
				SecretName:     "cool_secret",
				Body:           literal_6879,
				RandomOverride: fakeRandom,
				Application:    tt.opts.Application,
			}

			err := setRun(opts)
			assert.NoError(t, err)

			reg.Verify(t)

			data, err := io.ReadAll(reg.Requests[1].Body)
			assert.NoError(t, err)
			var payload SecretPayload
			err = json.Unmarshal(data, &payload)
			assert.NoError(t, err)
			assert.Equal(t, payload.KeyID, "123")
			assert.Equal(t, payload.EncryptedValue, literal_4276)
		})
	}
}

func TestSetRunenv(t *testing.T) {
	reg := &httpmock.Registry{}

	reg.Register(httpmock.REST("GET", "repos/owner/repo/environments/development/secrets/public-key"),
		httpmock.JSONResponse(PubKey{ID: "123", Key: literal_9163}))

	reg.Register(httpmock.REST("PUT", "repos/owner/repo/environments/development/secrets/cool_secret"), httpmock.StatusStringResponse(201, `{}`))

	ios, _, _, _ := iostreams.Test()

	opts := &SetOptions{
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		Config: func() (config.Config, error) { return config.NewBlankConfig(), nil },
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.FromFullName(literal_1049)
		},
		EnvName:        "development",
		IO:             ios,
		SecretName:     "cool_secret",
		Body:           literal_6879,
		RandomOverride: fakeRandom,
	}

	err := setRun(opts)
	assert.NoError(t, err)

	reg.Verify(t)

	data, err := io.ReadAll(reg.Requests[1].Body)
	assert.NoError(t, err)
	var payload SecretPayload
	err = json.Unmarshal(data, &payload)
	assert.NoError(t, err)
	assert.Equal(t, payload.KeyID, "123")
	assert.Equal(t, payload.EncryptedValue, literal_4276)
}

func TestSetRunorg(t *testing.T) {
	tests := []struct {
		name                       string
		opts                       *SetOptions
		wantVisibility             shared.Visibility
		wantRepositories           []int64
		wantDependabotRepositories []string
		wantApp                    string
	}{
		{
			name: "all vis",
			opts: &SetOptions{
				OrgName:    "UmbrellaCorporation",
				Visibility: shared.All,
			},
			wantApp: "actions",
		},
		{
			name: "selected visibility",
			opts: &SetOptions{
				OrgName:         "UmbrellaCorporation",
				Visibility:      shared.Selected,
				RepositoryNames: []string{"birkin", "UmbrellaCorporation/wesker"},
			},
			wantRepositories: []int64{1, 2},
			wantApp:          "actions",
		},
		{
			name: "Dependabot",
			opts: &SetOptions{
				OrgName:     "UmbrellaCorporation",
				Visibility:  shared.All,
				Application: shared.Dependabot,
			},
			wantApp: "dependabot",
		},
		{
			name: "Dependabot selected visibility",
			opts: &SetOptions{
				OrgName:         "UmbrellaCorporation",
				Visibility:      shared.Selected,
				Application:     shared.Dependabot,
				RepositoryNames: []string{"birkin", "UmbrellaCorporation/wesker"},
			},
			wantDependabotRepositories: []string{"1", "2"},
			wantApp:                    "dependabot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}

			orgName := tt.opts.OrgName

			reg.Register(httpmock.REST("GET",
				fmt.Sprintf("orgs/%s/%s/secrets/public-key", orgName, tt.wantApp)),
				httpmock.JSONResponse(PubKey{ID: "123", Key: literal_9163}))

			reg.Register(httpmock.REST("PUT",
				fmt.Sprintf("orgs/%s/%s/secrets/cool_secret", orgName, tt.wantApp)),
				httpmock.StatusStringResponse(201, `{}`))

			if len(tt.opts.RepositoryNames) > 0 {
				reg.Register(httpmock.GraphQL(`query MapRepositoryNames\b`),
					httpmock.StringResponse(`{"data":{"repo_0001":{"databaseId":1},"repo_0002":{"databaseId":2}}}`))
			}

			ios, _, _, _ := iostreams.Test()

			tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.FromFullName(literal_1049)
			}
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}
			tt.opts.Config = func() (config.Config, error) {
				return config.NewBlankConfig(), nil
			}
			tt.opts.IO = ios
			tt.opts.SecretName = "cool_secret"
			tt.opts.Body = literal_6879
			tt.opts.RandomOverride = fakeRandom

			err := setRun(tt.opts)
			assert.NoError(t, err)

			reg.Verify(t)

			data, err := io.ReadAll(reg.Requests[len(reg.Requests)-1].Body)
			assert.NoError(t, err)

			if tt.opts.Application == shared.Dependabot {
				var payload DependabotSecretPayload
				err = json.Unmarshal(data, &payload)
				assert.NoError(t, err)
				assert.Equal(t, payload.KeyID, "123")
				assert.Equal(t, payload.EncryptedValue, literal_4276)
				assert.Equal(t, payload.Visibility, tt.opts.Visibility)
				assert.ElementsMatch(t, payload.Repositories, tt.wantDependabotRepositories)
			} else {
				var payload SecretPayload
				err = json.Unmarshal(data, &payload)
				assert.NoError(t, err)
				assert.Equal(t, payload.KeyID, "123")
				assert.Equal(t, payload.EncryptedValue, literal_4276)
				assert.Equal(t, payload.Visibility, tt.opts.Visibility)
				assert.ElementsMatch(t, payload.Repositories, tt.wantRepositories)
			}
		})
	}
}

func TestSetRunuser(t *testing.T) {
	tests := []struct {
		name             string
		opts             *SetOptions
		wantVisibility   shared.Visibility
		wantRepositories []int64
	}{
		{
			name: "all vis",
			opts: &SetOptions{
				UserSecrets: true,
				Visibility:  shared.All,
			},
		},
		{
			name: "selected visibility",
			opts: &SetOptions{
				UserSecrets:     true,
				Visibility:      shared.Selected,
				RepositoryNames: []string{literal_5461, "github/hub"},
			},
			wantRepositories: []int64{212613049, 401025},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}

			reg.Register(httpmock.REST("GET", "user/codespaces/secrets/public-key"),
				httpmock.JSONResponse(PubKey{ID: "123", Key: literal_9163}))

			reg.Register(httpmock.REST("PUT", "user/codespaces/secrets/cool_secret"),
				httpmock.StatusStringResponse(201, `{}`))

			if len(tt.opts.RepositoryNames) > 0 {
				reg.Register(httpmock.GraphQL(`query MapRepositoryNames\b`),
					httpmock.StringResponse(`{"data":{"repo_0001":{"databaseId":212613049},"repo_0002":{"databaseId":401025}}}`))
			}

			ios, _, _, _ := iostreams.Test()

			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}
			tt.opts.Config = func() (config.Config, error) {
				return config.NewBlankConfig(), nil
			}
			tt.opts.IO = ios
			tt.opts.SecretName = "cool_secret"
			tt.opts.Body = literal_6879
			tt.opts.RandomOverride = fakeRandom

			err := setRun(tt.opts)
			assert.NoError(t, err)

			reg.Verify(t)

			data, err := io.ReadAll(reg.Requests[len(reg.Requests)-1].Body)
			assert.NoError(t, err)
			var payload SecretPayload
			err = json.Unmarshal(data, &payload)
			assert.NoError(t, err)
			assert.Equal(t, payload.KeyID, "123")
			assert.Equal(t, payload.EncryptedValue, literal_4276)
			assert.ElementsMatch(t, payload.Repositories, tt.wantRepositories)
		})
	}
}

func TestSetRunshouldNotStore(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(httpmock.REST("GET", "repos/owner/repo/actions/secrets/public-key"),
		httpmock.JSONResponse(PubKey{ID: "123", Key: literal_9163}))

	ios, _, stdout, stderr := iostreams.Test()

	opts := &SetOptions{
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		Config: func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		},
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.FromFullName(literal_1049)
		},
		IO:             ios,
		Body:           literal_6879,
		DoNotStore:     true,
		RandomOverride: fakeRandom,
	}

	err := setRun(opts)
	assert.NoError(t, err)

	assert.Equal(t, "UKYUCbHd0DJemxa3AOcZ6XcsBwALG9d4bpB8ZT0gSV39vl3BHiGSgj8zJapDxgB2BwqNqRhpjC4=\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestGetBody(t *testing.T) {
	tests := []struct {
		name    string
		bodyArg string
		want    string
		stdin   string
	}{
		{
			name:    "literal value",
			bodyArg: literal_6879,
			want:    literal_6879,
		},
		{
			name:  "from stdin",
			want:  literal_6879,
			stdin: literal_6879,
		},
		{
			name:  "from stdin with trailing newline character",
			want:  literal_6879,
			stdin: "a secret\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, _, _ := iostreams.Test()

			ios.SetStdinTTY(false)

			_, err := stdin.WriteString(tt.stdin)
			assert.NoError(t, err)

			body, err := getBody(&SetOptions{
				Body: tt.bodyArg,
				IO:   ios,
			})
			assert.NoError(t, err)

			assert.Equal(t, tt.want, string(body))
		})
	}
}

func TestGetBodyPrompt(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.SetStdinTTY(true)
	ios.SetStdoutTTY(true)

	pm := prompter.NewMockPrompter(t)
	pm.RegisterPassword("Paste your secret:", func(_ string) (string, error) {
		return "cool secret", nil
	})

	body, err := getBody(&SetOptions{
		IO:       ios,
		Prompter: pm,
	})
	assert.NoError(t, err)
	assert.Equal(t, string(body), "cool secret")
}

func TestGetSecretsFromOptions(t *testing.T) {
	genFile := func(s string) string {
		f, err := os.CreateTemp("", "gh-env.*")
		if err != nil {
			t.Fatal(err)
			return ""
		}
		defer f.Close()
		t.Cleanup(func() {
			_ = os.Remove(f.Name())
		})
		_, err = f.WriteString(s)
		if err != nil {
			t.Fatal(err)
		}
		return f.Name()
	}

	tests := []struct {
		name    string
		opts    SetOptions
		isTTY   bool
		stdin   string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "secret from arg",
			opts: SetOptions{
				SecretName: "FOO",
				Body:       "bar",
				EnvFile:    "",
			},
			want: map[string]string{"FOO": "bar"},
		},
		{
			name: "secrets from stdin",
			opts: SetOptions{
				Body:    "",
				EnvFile: "-",
			},
			stdin: `FOO=bar`,
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name: "secrets from file",
			opts: SetOptions{
				Body: "",
				EnvFile: genFile(heredoc.Doc(`
					FOO=bar
					QUOTED="my value"
					#IGNORED=true
					export SHELL=bash
				`)),
			},
			stdin: `FOO=bar`,
			want: map[string]string{
				"FOO":    "bar",
				"SHELL":  "bash",
				"QUOTED": "my value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, _, _ := iostreams.Test()
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStdoutTTY(tt.isTTY)
			stdin.WriteString(tt.stdin)
			opts := tt.opts
			opts.IO = ios
			gotSecrets, err := getSecretsFromOptions(&opts)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("getSecretsFromOptions() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if tt.wantErr {
				t.Fatalf("getSecretsFromOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(gotSecrets) != len(tt.want) {
				t.Fatalf("getSecretsFromOptions() = got %d secrets, want %d", len(gotSecrets), len(tt.want))
			}
			for k, v := range gotSecrets {
				if tt.want[k] != string(v) {
					t.Errorf("getSecretsFromOptions() %s = got %q, want %q", k, string(v), tt.want[k])
				}
			}
		})
	}
}

func fakeRandom() io.Reader {
	return bytes.NewReader(bytes.Repeat([]byte{5}, 32))
}

const literal_5461 = "cli/cli"

const literal_6879 = "a secret"

const literal_9163 = "CDjXqf7AJBXWhMczcy+Fs7JlACEptgceysutztHaFQI="

const literal_1049 = "owner/repo"

const literal_4276 = "UKYUCbHd0DJemxa3AOcZ6XcsBwALG9d4bpB8ZT0gSV39vl3BHiGSgj8zJapDxgB2BwqNqRhpjC4="
