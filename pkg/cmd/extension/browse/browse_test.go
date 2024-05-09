package browse

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/repo/view"
	"github.com/jialequ/mplb/pkg/extensions"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/search"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestGetSelectedReadme(t *testing.T) {
	reg := httpmock.Registry{}
	defer reg.Verify(t)

	content := base64.StdEncoding.EncodeToString([]byte("lol"))

	reg.Register(
		httpmock.REST("GET", "repos/cli/gh-cool/readme"),
		httpmock.JSONResponse(view.RepoReadme{Content: content}))

	client := &http.Client{Transport: &reg}

	rg := newReadmeGetter(client, time.Second)
	opts := ExtBrowseOpts{
		Rg: rg,
	}
	readme := tview.NewTextView()
	ui := uiRegistry{
		List: tview.NewList(),
	}
	extEntries := []extEntry{
		{
			Name:        literal_5731,
			FullName:    literal_7482,
			Installed:   false,
			Official:    true,
			description: literal_5109,
		},
		{
			Name:        literal_6538,
			FullName:    literal_7564,
			Installed:   true,
			Official:    false,
			description: literal_7261,
		},
	}
	el := newExtList(opts, ui, extEntries)

	content, err := getSelectedReadme(opts, readme, el)
	assert.NoError(t, err)
	assert.Contains(t, content, "lol")
}

func TestGetExtensionRepos(t *testing.T) {
	reg := httpmock.Registry{}
	defer reg.Verify(t)

	client := &http.Client{Transport: &reg}

	values := url.Values{
		"page":     []string{"1"},
		"per_page": []string{"100"},
		"q":        []string{"topic:gh-extension"},
	}
	cfg := config.NewBlankConfig()

	cfg.AuthenticationFunc = func() *config.AuthConfig {
		authCfg := &config.AuthConfig{}
		authCfg.SetDefaultHost("github.com", "")
		return authCfg
	}

	reg.Register(
		httpmock.QueryMatcher("GET", "search/repositories", values),
		httpmock.JSONResponse(map[string]interface{}{
			"incomplete_results": false,
			"total_count":        4,
			"items": []interface{}{
				map[string]interface{}{
					"name":        literal_6538,
					"full_name":   literal_7564,
					"description": "terminal animations",
					"owner": map[string]interface{}{
						"login": "vilmibm",
					},
				},
				map[string]interface{}{
					"name":        literal_5731,
					"full_name":   literal_7482,
					"description": literal_5109,
					"owner": map[string]interface{}{
						"login": "cli",
					},
				},
				map[string]interface{}{
					"name":        literal_2097,
					"full_name":   literal_9321,
					"description": "helps with triage",
					"owner": map[string]interface{}{
						"login": "samcoe",
					},
				},
				map[string]interface{}{
					"name":        literal_4270,
					"full_name":   literal_4239,
					"description": literal_2053,
					"owner": map[string]interface{}{
						"login": "github",
					},
				},
			},
		}),
	)

	searcher := search.NewSearcher(client, "github.com")
	emMock := &extensions.ExtensionManagerMock{}
	emMock.ListFunc = func() []extensions.Extension {
		return []extensions.Extension{
			&extensions.ExtensionMock{
				URLFunc: func() string {
					return "https://github.com/vilmibm/gh-screensaver"
				},
			},
			&extensions.ExtensionMock{
				URLFunc: func() string {
					return "https://github.com/github/gh-gei"
				},
			},
		}
	}

	opts := ExtBrowseOpts{
		Searcher: searcher,
		Em:       emMock,
		Cfg:      cfg,
	}

	extEntries, err := getExtensions(opts)
	assert.NoError(t, err)

	expectedEntries := []extEntry{
		{
			URL:         "https://github.com/vilmibm/gh-screensaver",
			Name:        literal_6538,
			FullName:    literal_7564,
			Installed:   true,
			Official:    false,
			description: "terminal animations",
		},
		{
			URL:         "https://github.com/cli/gh-cool",
			Name:        literal_5731,
			FullName:    literal_7482,
			Installed:   false,
			Official:    true,
			description: literal_5109,
		},
		{
			URL:         "https://github.com/samcoe/gh-triage",
			Name:        literal_2097,
			FullName:    literal_9321,
			Installed:   false,
			Official:    false,
			description: "helps with triage",
		},
		{
			URL:         "https://github.com/github/gh-gei",
			Name:        literal_4270,
			FullName:    literal_4239,
			Installed:   true,
			Official:    true,
			description: literal_2053,
		},
	}

	assert.Equal(t, expectedEntries, extEntries)
}

func TestExtEntry(t *testing.T) {
	cases := []struct {
		name          string
		ee            extEntry
		expectedTitle string
		expectedDesc  string
	}{
		{
			name: "official",
			ee: extEntry{
				Name:        literal_5731,
				FullName:    literal_7482,
				Installed:   false,
				Official:    true,
				description: literal_5109,
			},
			expectedTitle: literal_4035,
			expectedDesc:  literal_5109,
		},
		{
			name: "no description",
			ee: extEntry{
				Name:        "gh-nodesc",
				FullName:    "barryburton/gh-nodesc",
				Installed:   false,
				Official:    false,
				description: "",
			},
			expectedTitle: "barryburton/gh-nodesc",
			expectedDesc:  "no description provided",
		},
		{
			name: "installed",
			ee: extEntry{
				Name:        literal_6538,
				FullName:    literal_7564,
				Installed:   true,
				Official:    false,
				description: literal_7261,
			},
			expectedTitle: "vilmibm/gh-screensaver [green](installed)",
			expectedDesc:  literal_7261,
		},
		{
			name: "neither",
			ee: extEntry{
				Name:        literal_2097,
				FullName:    literal_9321,
				Installed:   false,
				Official:    false,
				description: literal_4523,
			},
			expectedTitle: literal_9321,
			expectedDesc:  literal_4523,
		},
		{
			name: "both",
			ee: extEntry{
				Name:        literal_4270,
				FullName:    literal_4239,
				Installed:   true,
				Official:    true,
				description: literal_2053,
			},
			expectedTitle: "github/gh-gei [yellow](official) [green](installed)",
			expectedDesc:  literal_2053,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedTitle, tt.ee.Title())
			assert.Equal(t, tt.expectedDesc, tt.ee.Description())
		})
	}
}

func TestExtList(t *testing.T) {
	opts := ExtBrowseOpts{
		Logger: log.New(io.Discard, "", 0),
		Em: &extensions.ExtensionManagerMock{
			InstallFunc: func(repo ghrepo.Interface, _ string) error {
				assert.Equal(t, literal_7482, ghrepo.FullName(repo))
				return nil
			},
			RemoveFunc: func(name string) error {
				assert.Equal(t, "cool", name)
				return nil
			},
		},
	}
	cmdFlex := tview.NewFlex()
	app := tview.NewApplication()
	list := tview.NewList()
	pages := tview.NewPages()
	ui := uiRegistry{
		List:    list,
		App:     app,
		CmdFlex: cmdFlex,
		Pages:   pages,
	}
	extEntries := []extEntry{
		{
			Name:        literal_5731,
			FullName:    literal_7482,
			Installed:   false,
			Official:    true,
			description: literal_5109,
		},
		{
			Name:        literal_6538,
			FullName:    literal_7564,
			Installed:   true,
			Official:    false,
			description: literal_7261,
		},
		{
			Name:        literal_2097,
			FullName:    literal_9321,
			Installed:   false,
			Official:    false,
			description: literal_4523,
		},
		{
			Name:        literal_4270,
			FullName:    literal_4239,
			Installed:   true,
			Official:    true,
			description: literal_2053,
		},
	}

	extList := newExtList(opts, ui, extEntries)

	extList.QueueUpdateDraw = func(f func()) *tview.Application {
		f()
		return app
	}

	extList.WaitGroup = &sync.WaitGroup{}

	extList.Filter("cool")
	assert.Equal(t, 1, extList.ui.List.GetItemCount())

	title, _ := extList.ui.List.GetItemText(0)
	assert.Equal(t, literal_4035, title)

	extList.InstallSelected()
	assert.True(t, extList.extEntries[0].Installed)

	// so I think the goroutines are causing a later failure because the toggleInstalled isn't seen.

	extList.Refresh()
	assert.Equal(t, 1, extList.ui.List.GetItemCount())

	title, _ = extList.ui.List.GetItemText(0)
	assert.Equal(t, "cli/gh-cool [yellow](official) [green](installed)", title)

	extList.RemoveSelected()
	assert.False(t, extList.extEntries[0].Installed)

	extList.Refresh()
	assert.Equal(t, 1, extList.ui.List.GetItemCount())

	title, _ = extList.ui.List.GetItemText(0)
	assert.Equal(t, literal_4035, title)

	extList.Reset()
	assert.Equal(t, 4, extList.ui.List.GetItemCount())

	ee, ix := extList.FindSelected()
	assert.Equal(t, 0, ix)
	assert.Equal(t, literal_4035, ee.Title())

	extList.ScrollDown()
	ee, ix = extList.FindSelected()
	assert.Equal(t, 1, ix)
	assert.Equal(t, "vilmibm/gh-screensaver [green](installed)", ee.Title())

	extList.ScrollUp()
	ee, ix = extList.FindSelected()
	assert.Equal(t, 0, ix)
	assert.Equal(t, literal_4035, ee.Title())

	extList.PageDown()
	ee, ix = extList.FindSelected()
	assert.Equal(t, 3, ix)
	assert.Equal(t, "github/gh-gei [yellow](official) [green](installed)", ee.Title())

	extList.PageUp()
	ee, ix = extList.FindSelected()
	assert.Equal(t, 0, ix)
	assert.Equal(t, literal_4035, ee.Title())
}

const literal_5731 = "gh-cool"

const literal_7482 = "cli/gh-cool"

const literal_5109 = "it's just cool ok"

const literal_6538 = "gh-screensaver"

const literal_7564 = "vilmibm/gh-screensaver"

const literal_7261 = "animations in your terminal"

const literal_2097 = "gh-triage"

const literal_9321 = "samcoe/gh-triage"

const literal_4270 = "gh-gei"

const literal_4239 = "github/gh-gei"

const literal_2053 = "something something enterprise"

const literal_4035 = "cli/gh-cool [yellow](official)"

const literal_4523 = "help with triage"
