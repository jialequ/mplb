package browse

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdBrowse(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		factory  func(*cmdutil.Factory) *cmdutil.Factory
		wants    BrowseOptions
		wantsErr bool
	}{
		{
			name:     "no arguments",
			cli:      "",
			wantsErr: false,
		},
		{
			name: "settings flag",
			cli:  "--settings",
			wants: BrowseOptions{
				SettingsFlag: true,
			},
			wantsErr: false,
		},
		{
			name: "projects flag",
			cli:  "--projects",
			wants: BrowseOptions{
				ProjectsFlag: true,
			},
			wantsErr: false,
		},
		{
			name: "releases flag",
			cli:  "--releases",
			wants: BrowseOptions{
				ReleasesFlag: true,
			},
			wantsErr: false,
		},
		{
			name: "wiki flag",
			cli:  "--wiki",
			wants: BrowseOptions{
				WikiFlag: true,
			},
			wantsErr: false,
		},
		{
			name: "no browser flag",
			cli:  "--no-browser",
			wants: BrowseOptions{
				NoBrowserFlag: true,
			},
			wantsErr: false,
		},
		{
			name: "branch flag",
			cli:  "--branch main",
			wants: BrowseOptions{
				Branch: "main",
			},
			wantsErr: false,
		},
		{
			name:     "branch flag without a branch name",
			cli:      "--branch",
			wantsErr: true,
		},
		{
			name: "combination: settings projects",
			cli:  "--settings --projects",
			wants: BrowseOptions{
				SettingsFlag: true,
				ProjectsFlag: true,
			},
			wantsErr: true,
		},
		{
			name: "combination: projects wiki",
			cli:  "--projects --wiki",
			wants: BrowseOptions{
				ProjectsFlag: true,
				WikiFlag:     true,
			},
			wantsErr: true,
		},
		{
			name: "passed argument",
			cli:  literal_2689,
			wants: BrowseOptions{
				SelectorArg: literal_2689,
			},
			wantsErr: false,
		},
		{
			name:     "passed two arguments",
			cli:      "main.go main.go",
			wantsErr: true,
		},
		{
			name:     "passed argument and projects flag",
			cli:      "main.go --projects",
			wantsErr: true,
		},
		{
			name:     "passed argument and releases flag",
			cli:      "main.go --releases",
			wantsErr: true,
		},
		{
			name:     "passed argument and settings flag",
			cli:      "main.go --settings",
			wantsErr: true,
		},
		{
			name:     "passed argument and wiki flag",
			cli:      "main.go --wiki",
			wantsErr: true,
		},
		{
			name: "empty commit flag",
			cli:  "--commit",
			wants: BrowseOptions{
				Commit: emptyCommitFlag,
			},
			wantsErr: false,
		},
		{
			name: "commit flag with a hash",
			cli:  "--commit=12a4",
			wants: BrowseOptions{
				Commit: "12a4",
			},
			wantsErr: false,
		},
		{
			name: "commit flag with a hash and a file selector",
			cli:  "main.go --commit=12a4",
			wants: BrowseOptions{
				Commit:      "12a4",
				SelectorArg: literal_2689,
			},
			wantsErr: false,
		},
		{
			name:     "passed both branch and commit flags",
			cli:      "main.go --branch main --commit=12a4",
			wantsErr: true,
		},
		{
			name:     "passed both number arg and branch flag",
			cli:      "1 --branch trunk",
			wantsErr: true,
		},
		{
			name:     "passed both number arg and commit flag",
			cli:      "1 --commit=12a4",
			wantsErr: true,
		},
		{
			name:     "passed both commit SHA arg and branch flag",
			cli:      "de07febc26e19000f8c9e821207f3bc34a3c8038 --branch trunk",
			wantsErr: true,
		},
		{
			name:     "passed both commit SHA arg and commit flag",
			cli:      "de07febc26e19000f8c9e821207f3bc34a3c8038 --commit=12a4",
			wantsErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := cmdutil.Factory{}
			var opts *BrowseOptions
			cmd := NewCmdBrowse(&f, func(o *BrowseOptions) error {
				opts = o
				return nil
			})
			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)
			cmd.SetArgs(argv)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			_, err = cmd.ExecuteC()

			if tt.wantsErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wants.Branch, opts.Branch)
			assert.Equal(t, tt.wants.SelectorArg, opts.SelectorArg)
			assert.Equal(t, tt.wants.ProjectsFlag, opts.ProjectsFlag)
			assert.Equal(t, tt.wants.ReleasesFlag, opts.ReleasesFlag)
			assert.Equal(t, tt.wants.WikiFlag, opts.WikiFlag)
			assert.Equal(t, tt.wants.NoBrowserFlag, opts.NoBrowserFlag)
			assert.Equal(t, tt.wants.SettingsFlag, opts.SettingsFlag)
			assert.Equal(t, tt.wants.Commit, opts.Commit)
		})
	}
}

func TestParsePathFromFileArg(t *testing.T) {
	tests := []struct {
		name         string
		currentDir   string
		fileArg      string
		expectedPath string
	}{
		{
			name:         "empty paths",
			currentDir:   "",
			fileArg:      "",
			expectedPath: "",
		},
		{
			name:         "root directory",
			currentDir:   "",
			fileArg:      ".",
			expectedPath: "",
		},
		{
			name:         "relative path",
			currentDir:   "",
			fileArg:      filepath.FromSlash("foo/bar.py"),
			expectedPath: "foo/bar.py",
		},
		{
			name:         "go to parent folder",
			currentDir:   literal_0437,
			fileArg:      filepath.FromSlash("../"),
			expectedPath: "pkg/cmd",
		},
		{
			name:         "current folder",
			currentDir:   literal_0437,
			fileArg:      ".",
			expectedPath: "pkg/cmd/browse",
		},
		{
			name:         "current folder (alternative)",
			currentDir:   literal_0437,
			fileArg:      filepath.FromSlash("./"),
			expectedPath: "pkg/cmd/browse",
		},
		{
			name:         "file that starts with '.'",
			currentDir:   literal_0437,
			fileArg:      ".gitignore",
			expectedPath: "pkg/cmd/browse/.gitignore",
		},
		{
			name:         "file in current folder",
			currentDir:   literal_0437,
			fileArg:      filepath.Join(".", "browse.go"),
			expectedPath: "pkg/cmd/browse/browse.go",
		},
		{
			name:         "file within parent folder",
			currentDir:   literal_0437,
			fileArg:      filepath.Join("..", "browse.go"),
			expectedPath: "pkg/cmd/browse.go",
		},
		{
			name:         "file within parent folder uncleaned",
			currentDir:   literal_0437,
			fileArg:      filepath.FromSlash(".././//browse.go"),
			expectedPath: "pkg/cmd/browse.go",
		},
		{
			name:         "different path from root directory",
			currentDir:   literal_0437,
			fileArg:      filepath.Join("..", "..", "..", "internal/build/build.go"),
			expectedPath: "internal/build/build.go",
		},
		{
			name:         "go out of repository",
			currentDir:   literal_0437,
			fileArg:      filepath.FromSlash("../../../../../../"),
			expectedPath: "",
		},
		{
			name:         "go to root of repository",
			currentDir:   literal_0437,
			fileArg:      filepath.Join("../../../"),
			expectedPath: "",
		},
		{
			name:         "empty fileArg",
			fileArg:      "",
			expectedPath: "",
		},
	}
	for _, tt := range tests {
		path, _, _, _ := parseFile(BrowseOptions{
			PathFromRepoRoot: func() string {
				return tt.currentDir
			}}, tt.fileArg)
		assert.Equal(t, tt.expectedPath, path, tt.name)
	}
}

const literal_2689 = "main.go"

const literal_0437 = "pkg/cmd/browse/"
