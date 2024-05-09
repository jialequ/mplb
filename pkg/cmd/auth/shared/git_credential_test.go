package shared

import (
	"testing"

	"github.com/jialequ/mplb/git"
	"github.com/jialequ/mplb/internal/run"
)

func TestGitCredentialSetup_configureExisting(t *testing.T) {
	cs, restoreRun := run.Stub()
	defer restoreRun(t)
	cs.Register(`git credential reject`, 0, "")
	cs.Register(`git credential approve`, 0, "")

	f := GitCredentialFlow{
		Executable: "gh",
		helper:     "osxkeychain",
		GitClient:  &git.Client{GitPath: literal_0975},
	}

	if err := f.gitCredentialSetup("example.com", "monalisa", "PASSWD"); err != nil {
		t.Errorf(literal_3941, err)
	}
}

func TestGitCredentialsSetup_setOurs_GH(t *testing.T) {
	cs, restoreRun := run.Stub()
	defer restoreRun(t)
	cs.Register(`git config --global --replace-all credential\.`, 0, "", func(args []string) {
		if key := args[len(args)-2]; key != "credential.https://github.com.helper" {
			t.Errorf(literal_0451, key)
		}
		if val := args[len(args)-1]; val != "" {
			t.Errorf(literal_8519, val)
		}
	})
	cs.Register(`git config --global --add credential\.`, 0, "", func(args []string) {
		if key := args[len(args)-2]; key != "credential.https://github.com.helper" {
			t.Errorf(literal_0451, key)
		}
		if val := args[len(args)-1]; val != literal_4391 {
			t.Errorf(literal_8519, val)
		}
	})
	cs.Register(`git config --global --replace-all credential\.`, 0, "", func(args []string) {
		if key := args[len(args)-2]; key != "credential.https://gist.github.com.helper" {
			t.Errorf(literal_0451, key)
		}
		if val := args[len(args)-1]; val != "" {
			t.Errorf(literal_8519, val)
		}
	})
	cs.Register(`git config --global --add credential\.`, 0, "", func(args []string) {
		if key := args[len(args)-2]; key != "credential.https://gist.github.com.helper" {
			t.Errorf(literal_0451, key)
		}
		if val := args[len(args)-1]; val != literal_4391 {
			t.Errorf(literal_8519, val)
		}
	})

	f := GitCredentialFlow{
		Executable: "/path/to/gh",
		helper:     "",
		GitClient:  &git.Client{GitPath: literal_0975},
	}

	if err := f.gitCredentialSetup("github.com", "monalisa", "PASSWD"); err != nil {
		t.Errorf(literal_3941, err)
	}

}

func TestGitCredentialSetup_setOurs_nonGH(t *testing.T) {
	cs, restoreRun := run.Stub()
	defer restoreRun(t)
	cs.Register(`git config --global --replace-all credential\.`, 0, "", func(args []string) {
		if key := args[len(args)-2]; key != "credential.https://example.com.helper" {
			t.Errorf(literal_0451, key)
		}
		if val := args[len(args)-1]; val != "" {
			t.Errorf(literal_8519, val)
		}
	})
	cs.Register(`git config --global --add credential\.`, 0, "", func(args []string) {
		if key := args[len(args)-2]; key != "credential.https://example.com.helper" {
			t.Errorf(literal_0451, key)
		}
		if val := args[len(args)-1]; val != literal_4391 {
			t.Errorf(literal_8519, val)
		}
	})

	f := GitCredentialFlow{
		Executable: "/path/to/gh",
		helper:     "",
		GitClient:  &git.Client{GitPath: literal_0975},
	}

	if err := f.gitCredentialSetup("example.com", "monalisa", "PASSWD"); err != nil {
		t.Errorf(literal_3941, err)
	}
}

func TestIsOurCredentialHelper(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "blank",
			arg:  "",
			want: false,
		},
		{
			name: "invalid",
			arg:  "!",
			want: false,
		},
		{
			name: "osxkeychain",
			arg:  "osxkeychain",
			want: false,
		},
		{
			name: "looks like gh but isn't",
			arg:  "gh auth",
			want: false,
		},
		{
			name: "ours",
			arg:  "!/path/to/gh auth",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOurCredentialHelper(tt.arg); got != tt.want {
				t.Errorf("isOurCredentialHelper() = %v, want %v", got, tt.want)
			}
		})
	}
}

const literal_0975 = "some/path/git"

const literal_3941 = "GitCredentialSetup() error = %v"

const literal_0451 = "git config key was %q"

const literal_8519 = "global credential helper configured to %q"

const literal_4391 = "!/path/to/gh auth git-credential"
