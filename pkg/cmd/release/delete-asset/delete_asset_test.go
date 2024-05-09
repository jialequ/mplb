package deleteasset

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmd/release/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdDeleteAsset(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		isTTY   bool
		want    DeleteAssetOptions
		wantErr string
	}{
		{
			name:  "tag and asset arguments",
			args:  "v1.2.3 test-asset",
			isTTY: true,
			want: DeleteAssetOptions{
				TagName:     literal_5791,
				SkipConfirm: false,
				AssetName:   literal_2495,
			},
		},
		{
			name:  "skip confirm",
			args:  "v1.2.3 test-asset -y",
			isTTY: true,
			want: DeleteAssetOptions{
				TagName:     literal_5791,
				SkipConfirm: true,
				AssetName:   literal_2495,
			},
		},
		{
			name:    "no arguments",
			args:    "",
			isTTY:   true,
			wantErr: "accepts 2 arg(s), received 0",
		},
		{
			name:    "one arguments",
			args:    literal_5791,
			isTTY:   true,
			wantErr: "accepts 2 arg(s), received 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var opts *DeleteAssetOptions
			cmd := NewCmdDeleteAsset(f, func(o *DeleteAssetOptions) error {
				opts = o
				return nil
			})

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.TagName, opts.TagName)
			assert.Equal(t, tt.want.SkipConfirm, opts.SkipConfirm)
			assert.Equal(t, tt.want.AssetName, opts.AssetName)
		})
	}
}

func TestDeleteAssetRun(t *testing.T) {
	tests := []struct {
		name          string
		isTTY         bool
		opts          DeleteAssetOptions
		prompterStubs func(*prompter.PrompterMock)
		wantErr       string
		wantStdout    string
		wantStderr    string
	}{
		{
			name:  "interactive confirm",
			isTTY: true,
			opts: DeleteAssetOptions{
				TagName:   literal_5791,
				AssetName: literal_2495,
			},
			prompterStubs: func(pm *prompter.PrompterMock) {
				pm.ConfirmFunc = func(p string, d bool) (bool, error) {
					if p == "Delete asset test-asset in release v1.2.3 in OWNER/REPO?" {
						return true, nil
					}
					return false, prompter.NoSuchPromptErr(p)
				}
			},
			wantStdout: ``,
			wantStderr: "✓ Deleted asset test-asset from release v1.2.3\n",
		},
		{
			name:  "skipping confirmation",
			isTTY: true,
			opts: DeleteAssetOptions{
				TagName:     literal_5791,
				SkipConfirm: true,
				AssetName:   literal_2495,
			},
			wantStdout: ``,
			wantStderr: "✓ Deleted asset test-asset from release v1.2.3\n",
		},
		{
			name:  "non-interactive",
			isTTY: false,
			opts: DeleteAssetOptions{
				TagName:     literal_5791,
				SkipConfirm: false,
				AssetName:   literal_2495,
			},
			wantStdout: ``,
			wantStderr: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			fakeHTTP := &httpmock.Registry{}
			defer fakeHTTP.Verify(t)
			shared.StubFetchRelease(t, fakeHTTP, "OWNER", "REPO", tt.opts.TagName, `{
				"tag_name": literal_5791,
				"draft": false,
				"url": "https://api.github.com/repos/OWNER/REPO/releases/23456",
				"assets": [
					{
						"url": "https://api.github.com/repos/OWNER/REPO/releases/assets/1",
						"id": 1,
						"name": literal_2495
					}
				]
			}`)
			fakeHTTP.Register(httpmock.REST("DELETE", "repos/OWNER/REPO/releases/assets/1"), httpmock.StatusStringResponse(204, ""))

			pm := &prompter.PrompterMock{}
			if tt.prompterStubs != nil {
				tt.prompterStubs(pm)
			}

			tt.opts.IO = ios
			tt.opts.Prompter = pm
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: fakeHTTP}, nil
			}
			tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.FromFullName("OWNER/REPO")
			}

			err := deleteAssetRun(&tt.opts)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}

const literal_5791 = "v1.2.3"

const literal_2495 = "test-asset"
