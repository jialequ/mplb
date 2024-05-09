package extension

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/run"
	"github.com/jialequ/mplb/pkg/extensions"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if err := func(args []string) error {
		fmt.Fprintf(os.Stdout, "%v\n", args)
		return nil
	}(os.Args[3:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func newTestManager(dir string, client *http.Client, gitClient gitClient, ios *iostreams.IOStreams) *Manager {
	return &Manager{
		dataDir:  func() string { return dir },
		lookPath: func(exe string) (string, error) { return exe, nil },
		findSh:   func() (string, error) { return "sh", nil },
		newCommand: func(exe string, args ...string) *exec.Cmd {
			args = append([]string{os.Args[0], "-test.run=TestHelperProcess", "--", exe}, args...)
			cmd := exec.Command(args[0], args[1:]...)
			if ios != nil {
				cmd.Stdout = ios.Out
				cmd.Stderr = ios.ErrOut
			}
			cmd.Env = []string{"GH_WANT_HELPER_PROCESS=1"}
			return cmd
		},
		config:    config.NewBlankConfig(),
		io:        ios,
		client:    client,
		gitClient: gitClient,
		platform: func() (string, string) {
			return "windows-amd64", ".exe"
		},
	}
}

func TestManager_List(t *testing.T) {
	tempDir := t.TempDir()
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_4682, literal_4682)))
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_9362, literal_9362)))

	assert.NoError(t, stubBinaryExtension(
		filepath.Join(tempDir, "extensions", literal_9176),
		binManifest{
			Owner: "owner",
			Name:  literal_9176,
			Host:  literal_8564,
			Tag:   literal_3427,
		}))

	dirOne := filepath.Join(tempDir, "extensions", literal_4682)
	dirTwo := filepath.Join(tempDir, "extensions", literal_9362)
	gc, gcOne, gcTwo := &mockGitClient{}, &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", dirOne).Return(gcOne).Once()
	gc.On("ForRepo", dirTwo).Return(gcTwo).Once()

	m := newTestManager(tempDir, nil, gc, nil)
	exts := m.List()

	assert.Equal(t, 3, len(exts))
	assert.Equal(t, "bin-ext", exts[0].Name())
	assert.Equal(t, "hello", exts[1].Name())
	assert.Equal(t, "two", exts[2].Name())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
	gcTwo.AssertExpectations(t)
}

func TestManager_list_includeMetadata(t *testing.T) {
	tempDir := t.TempDir()

	assert.NoError(t, stubBinaryExtension(
		filepath.Join(tempDir, "extensions", literal_9176),
		binManifest{
			Owner: "owner",
			Name:  literal_9176,
			Host:  literal_8564,
			Tag:   literal_3427,
		}))

	reg := httpmock.Registry{}
	defer reg.Verify(t)
	client := http.Client{Transport: &reg}

	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Tag: literal_6327,
				Assets: []releaseAsset{
					{
						Name:   "gh-bin-ext-windows-amd64",
						APIURL: literal_4508,
					},
				},
			}))

	m := newTestManager(tempDir, &client, nil, nil)

	exts, err := m.list(true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	assert.Equal(t, "bin-ext", exts[0].Name())
	assert.True(t, exts[0].UpdateAvailable())
	assert.Equal(t, "https://example.com/owner/gh-bin-ext", exts[0].URL())
}

func TestManager_Remove(t *testing.T) {
	tempDir := t.TempDir()
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_4682, literal_4682)))
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_9362, literal_9362)))

	m := newTestManager(tempDir, nil, nil, nil)
	err := m.Remove("hello")
	assert.NoError(t, err)

	items, err := os.ReadDir(filepath.Join(tempDir, "extensions"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, literal_9362, items[0].Name())
}

func TestManager_Upgrade_NoExtensions(t *testing.T) {
	tempDir := t.TempDir()
	ios, _, stdout, stderr := iostreams.Test()
	m := newTestManager(tempDir, nil, nil, ios)
	err := m.Upgrade("", false)
	assert.EqualError(t, err, "no extensions installed")
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_Upgrade_NoMatchingExtension(t *testing.T) {
	tempDir := t.TempDir()
	extDir := filepath.Join(tempDir, "extensions", literal_4682)
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_4682, literal_4682)))
	ios, _, stdout, stderr := iostreams.Test()
	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", extDir).Return(gcOne).Once()
	m := newTestManager(tempDir, nil, gc, ios)
	err := m.Upgrade("invalid", false)
	assert.EqualError(t, err, `no extension matched "invalid"`)
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_UpgradeExtensions(t *testing.T) {
	tempDir := t.TempDir()
	dirOne := filepath.Join(tempDir, "extensions", literal_4682)
	dirTwo := filepath.Join(tempDir, "extensions", literal_9362)
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_4682, literal_4682)))
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_9362, literal_9362)))
	assert.NoError(t, stubLocalExtension(tempDir, filepath.Join(tempDir, "extensions", literal_9374, literal_9374)))
	ios, _, stdout, stderr := iostreams.Test()
	gc, gcOne, gcTwo := &mockGitClient{}, &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", dirOne).Return(gcOne).Times(3)
	gc.On("ForRepo", dirTwo).Return(gcTwo).Times(3)
	gcOne.On("Remotes").Return(nil, nil).Once()
	gcTwo.On("Remotes").Return(nil, nil).Once()
	gcOne.On("Pull", "", "").Return(nil).Once()
	gcTwo.On("Pull", "", "").Return(nil).Once()
	m := newTestManager(tempDir, nil, gc, ios)
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(exts))
	for i := 0; i < 3; i++ {
		exts[i].currentVersion = literal_3054
		exts[i].latestVersion = literal_4652
	}
	err = m.upgradeExtensions(exts, false)
	assert.NoError(t, err)
	assert.Equal(t, heredoc.Doc(
		`
		[hello]: upgraded from old vers to new vers
		[local]: local extensions can not be upgraded
		[two]: upgraded from old vers to new vers
		`,
	), stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
	gcTwo.AssertExpectations(t)
}

func TestManager_UpgradeExtensions_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	dirOne := filepath.Join(tempDir, "extensions", literal_4682)
	dirTwo := filepath.Join(tempDir, "extensions", literal_9362)
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_4682, literal_4682)))
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_9362, literal_9362)))
	assert.NoError(t, stubLocalExtension(tempDir, filepath.Join(tempDir, "extensions", literal_9374, literal_9374)))
	ios, _, stdout, stderr := iostreams.Test()
	gc, gcOne, gcTwo := &mockGitClient{}, &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", dirOne).Return(gcOne).Twice()
	gc.On("ForRepo", dirTwo).Return(gcTwo).Twice()
	gcOne.On("Remotes").Return(nil, nil).Once()
	gcTwo.On("Remotes").Return(nil, nil).Once()
	m := newTestManager(tempDir, nil, gc, ios)
	m.EnableDryRunMode()
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(exts))
	for i := 0; i < 3; i++ {
		exts[i].currentVersion = fmt.Sprintf("%d", i)
		exts[i].latestVersion = fmt.Sprintf("%d", i+1)
	}
	err = m.upgradeExtensions(exts, false)
	assert.NoError(t, err)
	assert.Equal(t, heredoc.Doc(
		`
 		[hello]: would have upgraded from 0 to 1
 		[local]: local extensions can not be upgraded
 		[two]: would have upgraded from 2 to 3
 		`,
	), stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
	gcTwo.AssertExpectations(t)
}

func TestManager_UpgradeExtension_LocalExtension(t *testing.T) {
	tempDir := t.TempDir()
	assert.NoError(t, stubLocalExtension(tempDir, filepath.Join(tempDir, "extensions", literal_9374, literal_9374)))

	ios, _, stdout, stderr := iostreams.Test()
	m := newTestManager(tempDir, nil, nil, ios)
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	err = m.upgradeExtension(exts[0], false)
	assert.EqualError(t, err, "local extensions can not be upgraded")
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_UpgradeExtension_LocalExtension_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	assert.NoError(t, stubLocalExtension(tempDir, filepath.Join(tempDir, "extensions", literal_9374, literal_9374)))

	ios, _, stdout, stderr := iostreams.Test()
	m := newTestManager(tempDir, nil, nil, ios)
	m.EnableDryRunMode()
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	err = m.upgradeExtension(exts[0], false)
	assert.EqualError(t, err, "local extensions can not be upgraded")
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_UpgradeExtension_GitExtension(t *testing.T) {
	tempDir := t.TempDir()
	extensionDir := filepath.Join(tempDir, "extensions", literal_1539)
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_1539, literal_1539)))
	ios, _, stdout, stderr := iostreams.Test()
	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", extensionDir).Return(gcOne).Times(3)
	gcOne.On("Remotes").Return(nil, nil).Once()
	gcOne.On("Pull", "", "").Return(nil).Once()
	m := newTestManager(tempDir, nil, gc, ios)
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.currentVersion = literal_3054
	ext.latestVersion = literal_4652
	err = m.upgradeExtension(ext, false)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_UpgradeExtension_GitExtension_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	extDir := filepath.Join(tempDir, "extensions", literal_1539)
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_1539, literal_1539)))
	ios, _, stdout, stderr := iostreams.Test()
	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", extDir).Return(gcOne).Twice()
	gcOne.On("Remotes").Return(nil, nil).Once()
	m := newTestManager(tempDir, nil, gc, ios)
	m.EnableDryRunMode()
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.currentVersion = literal_3054
	ext.latestVersion = literal_4652
	err = m.upgradeExtension(ext, false)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_UpgradeExtension_GitExtension_Force(t *testing.T) {
	tempDir := t.TempDir()
	extensionDir := filepath.Join(tempDir, "extensions", literal_1539)
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_1539, literal_1539)))
	ios, _, stdout, stderr := iostreams.Test()
	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", extensionDir).Return(gcOne).Times(3)
	gcOne.On("Remotes").Return(nil, nil).Once()
	gcOne.On("Fetch", "origin", "HEAD").Return(nil).Once()
	gcOne.On("CommandOutput", []string{"reset", "--hard", "origin/HEAD"}).Return("", nil).Once()
	m := newTestManager(tempDir, nil, gc, ios)
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.currentVersion = literal_3054
	ext.latestVersion = literal_4652
	err = m.upgradeExtension(ext, true)
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_MigrateToBinaryExtension(t *testing.T) {
	tempDir := t.TempDir()
	assert.NoError(t, stubExtension(filepath.Join(tempDir, "extensions", literal_1539, literal_1539)))
	ios, _, stdout, stderr := iostreams.Test()

	reg := httpmock.Registry{}
	defer reg.Verify(t)
	client := http.Client{Transport: &reg}
	gc := &gitExecuter{client: &git.Client{}}
	m := newTestManager(tempDir, &client, gc, ios)
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.currentVersion = literal_3054
	ext.latestVersion = literal_4652

	rs, restoreRun := run.Stub()
	defer restoreRun(t)

	rs.Register(`git -C.*?gh-remote remote -v`, 0, "origin  git@github.com:owner/gh-remote.git (fetch)\norigin  git@github.com:owner/gh-remote.git (push)")
	rs.Register(`git -C.*?gh-remote config --get-regexp \^.*`, 0, "remote.origin.gh-resolve base")

	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-remote/releases/latest"),
		httpmock.JSONResponse(
			release{
				Tag: literal_6327,
				Assets: []releaseAsset{
					{
						Name:   "gh-remote-windows-amd64.exe",
						APIURL: "/release/cool",
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-remote/releases/latest"),
		httpmock.JSONResponse(
			release{
				Tag: literal_6327,
				Assets: []releaseAsset{
					{
						Name:   "gh-remote-windows-amd64.exe",
						APIURL: "/release/cool",
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", literal_4169),
		httpmock.StringResponse(literal_7048))

	err = m.upgradeExtension(ext, false)
	assert.NoError(t, err)

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())

	manifest, err := os.ReadFile(filepath.Join(tempDir, "extensions/gh-remote", manifestName))
	assert.NoError(t, err)

	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	assert.NoError(t, err)

	assert.Equal(t, binManifest{
		Name:  literal_1539,
		Owner: "owner",
		Host:  "github.com",
		Tag:   literal_6327,
		Path:  filepath.Join(tempDir, "extensions/gh-remote/gh-remote.exe"),
	}, bm)

	fakeBin, err := os.ReadFile(filepath.Join(tempDir, "extensions/gh-remote/gh-remote.exe"))
	assert.NoError(t, err)

	assert.Equal(t, literal_7048, string(fakeBin))
}

func TestManager_UpgradeExtension_BinaryExtension(t *testing.T) {
	tempDir := t.TempDir()

	reg := httpmock.Registry{}
	defer reg.Verify(t)

	assert.NoError(t, stubBinaryExtension(
		filepath.Join(tempDir, "extensions", literal_9176),
		binManifest{
			Owner: "owner",
			Name:  literal_9176,
			Host:  literal_8564,
			Tag:   literal_3427,
		}))

	ios, _, stdout, stderr := iostreams.Test()
	m := newTestManager(tempDir, &http.Client{Transport: &reg}, nil, ios)
	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Tag: literal_6327,
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4508,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", "release/cool2"),
		httpmock.StringResponse(literal_7048))

	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.latestVersion = literal_6327
	err = m.upgradeExtension(ext, false)
	assert.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(tempDir, literal_9512, manifestName))
	assert.NoError(t, err)

	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	assert.NoError(t, err)

	assert.Equal(t, binManifest{
		Name:  literal_9176,
		Owner: "owner",
		Host:  literal_8564,
		Tag:   literal_6327,
		Path:  filepath.Join(tempDir, literal_1927),
	}, bm)

	fakeBin, err := os.ReadFile(filepath.Join(tempDir, literal_1927))
	assert.NoError(t, err)
	assert.Equal(t, literal_7048, string(fakeBin))

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_UpgradeExtension_BinaryExtension_Pinned_Force(t *testing.T) {
	tempDir := t.TempDir()

	reg := httpmock.Registry{}
	defer reg.Verify(t)

	assert.NoError(t, stubBinaryExtension(
		filepath.Join(tempDir, "extensions", literal_9176),
		binManifest{
			Owner:    "owner",
			Name:     literal_9176,
			Host:     literal_8564,
			Tag:      literal_3427,
			IsPinned: true,
		}))

	ios, _, stdout, stderr := iostreams.Test()
	m := newTestManager(tempDir, &http.Client{Transport: &reg}, nil, ios)
	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Tag: literal_6327,
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4508,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", "release/cool2"),
		httpmock.StringResponse(literal_7048))

	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.latestVersion = literal_6327
	err = m.upgradeExtension(ext, true)
	assert.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(tempDir, literal_9512, manifestName))
	assert.NoError(t, err)

	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	assert.NoError(t, err)

	assert.Equal(t, binManifest{
		Name:  literal_9176,
		Owner: "owner",
		Host:  literal_8564,
		Tag:   literal_6327,
		Path:  filepath.Join(tempDir, literal_1927),
	}, bm)

	fakeBin, err := os.ReadFile(filepath.Join(tempDir, literal_1927))
	assert.NoError(t, err)
	assert.Equal(t, literal_7048, string(fakeBin))

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_UpgradeExtension_BinaryExtension_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	reg := httpmock.Registry{}
	defer reg.Verify(t)
	assert.NoError(t, stubBinaryExtension(
		filepath.Join(tempDir, "extensions", literal_9176),
		binManifest{
			Owner: "owner",
			Name:  literal_9176,
			Host:  literal_8564,
			Tag:   literal_3427,
		}))

	ios, _, stdout, stderr := iostreams.Test()
	m := newTestManager(tempDir, &http.Client{Transport: &reg}, nil, ios)
	m.EnableDryRunMode()
	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Tag: literal_6327,
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4508,
					},
				},
			}))
	exts, err := m.list(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	ext.latestVersion = literal_6327
	err = m.upgradeExtension(ext, false)
	assert.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(tempDir, literal_9512, manifestName))
	assert.NoError(t, err)

	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	assert.NoError(t, err)

	assert.Equal(t, binManifest{
		Name:  literal_9176,
		Owner: "owner",
		Host:  literal_8564,
		Tag:   literal_3427,
	}, bm)
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_UpgradeExtension_BinaryExtension_Pinned(t *testing.T) {
	tempDir := t.TempDir()

	assert.NoError(t, stubBinaryExtension(
		filepath.Join(tempDir, "extensions", literal_9176),
		binManifest{
			Owner:    "owner",
			Name:     literal_9176,
			Host:     literal_8564,
			Tag:      "v1.6.3",
			IsPinned: true,
		}))

	ios, _, _, _ := iostreams.Test()
	m := newTestManager(tempDir, nil, nil, ios)
	exts, err := m.list(false)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]

	err = m.upgradeExtension(ext, false)
	assert.NotNil(t, err)
	assert.Equal(t, err, pinnedExtensionUpgradeError)
}

func TestManager_UpgradeExtension_GitExtension_Pinned(t *testing.T) {
	tempDir := t.TempDir()
	extDir := filepath.Join(tempDir, "extensions", literal_1539)
	assert.NoError(t, stubPinnedExtension(filepath.Join(extDir, literal_1539), "abcd1234"))

	ios, _, _, _ := iostreams.Test()

	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", extDir).Return(gcOne).Once()

	m := newTestManager(tempDir, nil, gc, ios)

	exts, err := m.list(false)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(exts))
	ext := exts[0]
	pinnedTrue := true
	ext.isPinned = &pinnedTrue
	ext.latestVersion = literal_4652

	err = m.upgradeExtension(ext, false)
	assert.NotNil(t, err)
	assert.Equal(t, err, pinnedExtensionUpgradeError)
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_Install_git(t *testing.T) {
	tempDir := t.TempDir()

	reg := httpmock.Registry{}
	defer reg.Verify(t)
	client := http.Client{Transport: &reg}

	ios, _, stdout, stderr := iostreams.Test()

	extensionDir := filepath.Join(tempDir, "extensions", "gh-some-ext")
	gc := &mockGitClient{}
	gc.On("Clone", "https://github.com/owner/gh-some-ext.git", []string{extensionDir}).Return("", nil).Once()

	m := newTestManager(tempDir, &client, gc, ios)

	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-some-ext/releases/latest"),
		httpmock.JSONResponse(
			release{
				Assets: []releaseAsset{
					{
						Name:   "not-a-binary",
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-some-ext/contents/gh-some-ext"),
		httpmock.StringResponse("script"))

	repo := ghrepo.New("owner", "gh-some-ext")

	err := m.Install(repo, "")
	assert.NoError(t, err)
	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
}

func TestManager_Install_git_pinned(t *testing.T) {
	tempDir := t.TempDir()

	reg := httpmock.Registry{}
	defer reg.Verify(t)
	client := http.Client{Transport: &reg}

	ios, _, stdout, stderr := iostreams.Test()

	extensionDir := filepath.Join(tempDir, "extensions", literal_5183)
	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", extensionDir).Return(gcOne).Once()
	gc.On("Clone", "https://github.com/owner/gh-cool-ext.git", []string{extensionDir}).Return("", nil).Once()
	gcOne.On("CheckoutBranch", "abcd1234").Return(nil).Once()

	m := newTestManager(tempDir, &client, gc, ios)

	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-cool-ext/releases/latest"),
		httpmock.JSONResponse(
			release{
				Assets: []releaseAsset{
					{
						Name:   "not-a-binary",
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-cool-ext/commits/some-ref"),
		httpmock.StringResponse("abcd1234"))
	reg.Register(
		httpmock.REST("GET", "repos/owner/gh-cool-ext/contents/gh-cool-ext"),
		httpmock.StringResponse("script"))

	_ = os.MkdirAll(filepath.Join(m.installDir(), literal_5183), 0700)
	repo := ghrepo.New("owner", literal_5183)
	err := m.Install(repo, "some-ref")
	assert.NoError(t, err)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "", stdout.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_Install_binary_pinned(t *testing.T) {
	repo := ghrepo.NewWithHost("owner", literal_9176, literal_8564)

	reg := httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", "api/v3/repos/owner/gh-bin-ext/releases/tags/v1.6.3-pre"),
		httpmock.JSONResponse(
			release{
				Tag: literal_9438,
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", literal_4169),
		httpmock.StringResponse(literal_3589))

	ios, _, stdout, stderr := iostreams.Test()
	tempDir := t.TempDir()

	m := newTestManager(tempDir, &http.Client{Transport: &reg}, nil, ios)

	err := m.Install(repo, literal_9438)
	assert.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(tempDir, literal_9512, manifestName))
	assert.NoError(t, err)

	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	assert.NoError(t, err)

	assert.Equal(t, binManifest{
		Name:     literal_9176,
		Owner:    "owner",
		Host:     literal_8564,
		Tag:      literal_9438,
		IsPinned: true,
		Path:     filepath.Join(tempDir, literal_1927),
	}, bm)

	fakeBin, err := os.ReadFile(filepath.Join(tempDir, literal_1927))
	assert.NoError(t, err)
	assert.Equal(t, literal_3589, string(fakeBin))

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())

}

func TestManager_Install_binary_unsupported(t *testing.T) {
	repo := ghrepo.NewWithHost("owner", literal_9176, literal_8564)

	reg := httpmock.Registry{}
	defer reg.Verify(t)
	client := http.Client{Transport: &reg}

	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Assets: []releaseAsset{
					{
						Name:   "gh-bin-ext-linux-amd64",
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Tag: literal_3427,
				Assets: []releaseAsset{
					{
						Name:   "gh-bin-ext-linux-amd64",
						APIURL: literal_4531,
					},
				},
			}))

	ios, _, stdout, stderr := iostreams.Test()
	tempDir := t.TempDir()

	m := newTestManager(tempDir, &client, nil, ios)

	err := m.Install(repo, "")
	assert.EqualError(t, err, "gh-bin-ext unsupported for windows-amd64. Open an issue: `gh issue create -R owner/gh-bin-ext -t'Support windows-amd64'`")

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_Install_binary(t *testing.T) {
	repo := ghrepo.NewWithHost("owner", literal_9176, literal_8564)

	reg := httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.JSONResponse(
			release{
				Tag: literal_3427,
				Assets: []releaseAsset{
					{
						Name:   literal_3925,
						APIURL: literal_4531,
					},
				},
			}))
	reg.Register(
		httpmock.REST("GET", literal_4169),
		httpmock.StringResponse(literal_3589))

	ios, _, stdout, stderr := iostreams.Test()
	tempDir := t.TempDir()

	m := newTestManager(tempDir, &http.Client{Transport: &reg}, nil, ios)

	err := m.Install(repo, "")
	assert.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(tempDir, literal_9512, manifestName))
	assert.NoError(t, err)

	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	assert.NoError(t, err)

	assert.Equal(t, binManifest{
		Name:  literal_9176,
		Owner: "owner",
		Host:  literal_8564,
		Tag:   literal_3427,
		Path:  filepath.Join(tempDir, literal_1927),
	}, bm)

	fakeBin, err := os.ReadFile(filepath.Join(tempDir, literal_1927))
	assert.NoError(t, err)
	assert.Equal(t, literal_3589, string(fakeBin))

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_repo_not_found(t *testing.T) {
	repo := ghrepo.NewWithHost("owner", literal_9176, literal_8564)

	reg := httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.REST("GET", literal_6358),
		httpmock.StatusStringResponse(404, `{}`))
	reg.Register(
		httpmock.REST("GET", "api/v3/repos/owner/gh-bin-ext"),
		httpmock.StatusStringResponse(404, `{}`))

	ios, _, stdout, stderr := iostreams.Test()
	tempDir := t.TempDir()

	m := newTestManager(tempDir, &http.Client{Transport: &reg}, nil, ios)

	if err := m.Install(repo, ""); err != repositoryNotFoundErr {
		t.Errorf("expected repositoryNotFoundErr, got: %v", err)
	}

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestManager_Create(t *testing.T) {
	chdirTemp(t)
	err := os.MkdirAll(literal_7215, 0755)
	assert.NoError(t, err)

	ios, _, stdout, stderr := iostreams.Test()

	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", literal_7215).Return(gcOne).Once()
	gc.On("CommandOutput", []string{"init", "--quiet", literal_7215}).Return("", nil).Once()
	gcOne.On("CommandOutput", []string{"add", literal_7215, "--chmod=+x"}).Return("", nil).Once()
	gcOne.On("CommandOutput", []string{"commit", "-m", "initial commit"}).Return("", nil).Once()

	m := newTestManager(".", nil, gc, ios)

	err = m.Create(literal_7215, extensions.GitTemplateType)
	assert.NoError(t, err)
	files, err := os.ReadDir(literal_7215)
	assert.NoError(t, err)
	assert.Equal(t, []string{literal_7215}, fileNames(files))

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

func TestManager_Create_other_binary(t *testing.T) {
	chdirTemp(t)
	err := os.MkdirAll(literal_7215, 0755)
	assert.NoError(t, err)

	ios, _, stdout, stderr := iostreams.Test()

	gc, gcOne := &mockGitClient{}, &mockGitClient{}
	gc.On("ForRepo", literal_7215).Return(gcOne).Once()
	gc.On("CommandOutput", []string{"init", "--quiet", literal_7215}).Return("", nil).Once()
	gcOne.On("CommandOutput", []string{"add", filepath.Join("script", "build.sh"), "--chmod=+x"}).Return("", nil).Once()
	gcOne.On("CommandOutput", []string{"add", "."}).Return("", nil).Once()
	gcOne.On("CommandOutput", []string{"commit", "-m", "initial commit"}).Return("", nil).Once()

	m := newTestManager(".", nil, gc, ios)

	err = m.Create(literal_7215, extensions.OtherBinTemplateType)
	assert.NoError(t, err)

	files, err := os.ReadDir(literal_7215)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	files, err = os.ReadDir(filepath.Join(literal_7215, ".github", "workflows"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"release.yml"}, fileNames(files))

	files, err = os.ReadDir(filepath.Join(literal_7215, "script"))
	assert.NoError(t, err)
	assert.Equal(t, []string{"build.sh"}, fileNames(files))

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())
	gc.AssertExpectations(t)
	gcOne.AssertExpectations(t)
}

// chdirTemp changes the current working directory to a temporary directory for the duration of the test.
func chdirTemp(t *testing.T) {
	oldWd, _ := os.Getwd()
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
}

func fileNames(files []os.DirEntry) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name()
	}
	sort.Strings(names)
	return names
}

func stubExtension(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	return f.Close()
}

func stubPinnedExtension(path string, pinnedVersion string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	f.Close()

	pinPath := filepath.Join(filepath.Dir(path), fmt.Sprintf(".pin-%s", pinnedVersion))
	f, err = os.OpenFile(pinPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	return f.Close()
}

func stubLocalExtension(tempDir, path string) error {
	extDir, err := os.MkdirTemp(tempDir, "local-ext")
	if err != nil {
		return err
	}
	extFile, err := os.OpenFile(filepath.Join(extDir, filepath.Base(path)), os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	if err := extFile.Close(); err != nil {
		return err
	}

	linkPath := filepath.Dir(path)
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(linkPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(extDir)
	if err != nil {
		return err
	}
	return f.Close()
}

// Given the path where an extension should be installed and a manifest struct, creates a fake binary extension on disk
func stubBinaryExtension(installPath string, bm binManifest) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}
	fakeBinaryPath := filepath.Join(installPath, filepath.Base(installPath))
	fb, err := os.OpenFile(fakeBinaryPath, os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	err = fb.Close()
	if err != nil {
		return err
	}

	bs, err := yaml.Marshal(bm)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	manifestPath := filepath.Join(installPath, manifestName)

	fm, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open manifest for writing: %w", err)
	}
	_, err = fm.Write(bs)
	if err != nil {
		return fmt.Errorf("failed write manifest file: %w", err)
	}

	return fm.Close()
}

const literal_4682 = "gh-hello"

const literal_9362 = "gh-two"

const literal_9176 = "gh-bin-ext"

const literal_8564 = "example.com"

const literal_3427 = "v1.0.1"

const literal_6358 = "api/v3/repos/owner/gh-bin-ext/releases/latest"

const literal_6327 = "v1.0.2"

const literal_4508 = "https://example.com/release/cool2"

const literal_9374 = "gh-local"

const literal_3054 = "old version"

const literal_4652 = "new version"

const literal_1539 = "gh-remote"

const literal_4169 = "release/cool"

const literal_7048 = "FAKE UPGRADED BINARY"

const literal_3925 = "gh-bin-ext-windows-amd64.exe"

const literal_9512 = "extensions/gh-bin-ext"

const literal_1927 = "extensions/gh-bin-ext/gh-bin-ext.exe"

const literal_4531 = "https://example.com/release/cool"

const literal_5183 = "gh-cool-ext"

const literal_9438 = "v1.6.3-pre"

const literal_3589 = "FAKE BINARY"

const literal_7215 = "gh-test"
