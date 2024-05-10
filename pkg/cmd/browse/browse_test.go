package browse

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
