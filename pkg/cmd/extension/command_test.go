package extension

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/jialequ/mplb/internal/browser"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/extensions"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdExtension(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	assert.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	tests := []struct {
		name          string
		args          []string
		managerStubs  func(em *extensions.ExtensionManagerMock) func(*testing.T)
		prompterStubs func(pm *prompter.PrompterMock)
		httpStubs     func(reg *httpmock.Registry)
		browseStubs   func(*browser.Stub) func(*testing.T)
		isTTY         bool
		wantErr       bool
		errMsg        string
		wantStdout    string
		wantStderr    string
	}{
		{
			name: "search for extensions",
			args: []string{"search"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{
						&extensions.ExtensionMock{
							URLFunc: func() string {
								return literal_5902
							},
						},
						&extensions.ExtensionMock{
							URLFunc: func() string {
								return literal_4230
							},
						},
					}
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				values := url.Values{
					"page":     []string{"1"},
					"per_page": []string{"30"},
					"q":        []string{literal_9853},
				}
				reg.Register(
					httpmock.QueryMatcher("GET", literal_0923, values),
					httpmock.JSONResponse(searchResults(4)),
				)
			},
			isTTY:      true,
			wantStdout: "Showing 4 of 4 extensions\n\n   REPO                    DESCRIPTION\n✓  vilmibm/gh-screensaver  terminal animations\n   cli/gh-cool             it's just cool ok\n   samcoe/gh-triage        helps with triage\n✓  github/gh-gei           something something enterprise\n",
		},
		{
			name: "search for extensions non-tty",
			args: []string{"search"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{
						&extensions.ExtensionMock{
							URLFunc: func() string {
								return literal_5902
							},
						},
						&extensions.ExtensionMock{
							URLFunc: func() string {
								return literal_4230
							},
						},
					}
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				values := url.Values{
					"page":     []string{"1"},
					"per_page": []string{"30"},
					"q":        []string{literal_9853},
				}
				reg.Register(
					httpmock.QueryMatcher("GET", literal_0923, values),
					httpmock.JSONResponse(searchResults(4)),
				)
			},
			wantStdout: "installed\tvilmibm/gh-screensaver\tterminal animations\n\tcli/gh-cool\tit's just cool ok\n\tsamcoe/gh-triage\thelps with triage\ninstalled\tgithub/gh-gei\tsomething something enterprise\n",
		},
		{
			name: "search for extensions with keywords",
			args: []string{"search", "screen"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{
						&extensions.ExtensionMock{
							URLFunc: func() string {
								return literal_5902
							},
						},
						&extensions.ExtensionMock{
							URLFunc: func() string {
								return literal_4230
							},
						},
					}
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				values := url.Values{
					"page":     []string{"1"},
					"per_page": []string{"30"},
					"q":        []string{"screen topic:gh-extension"},
				}
				results := searchResults(1)
				reg.Register(
					httpmock.QueryMatcher("GET", literal_0923, values),
					httpmock.JSONResponse(results),
				)
			},
			wantStdout: "installed\tvilmibm/gh-screensaver\tterminal animations\n",
		},
		{
			name: "search for extensions with parameter flags",
			args: []string{"search", "--limit", "1", "--order", "asc", "--sort", "stars"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{}
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				values := url.Values{
					"page":     []string{"1"},
					"order":    []string{"asc"},
					"sort":     []string{"stars"},
					"per_page": []string{"1"},
					"q":        []string{literal_9853},
				}
				results := searchResults(1)
				reg.Register(
					httpmock.QueryMatcher("GET", literal_0923, values),
					httpmock.JSONResponse(results),
				)
			},
			wantStdout: "\tvilmibm/gh-screensaver\tterminal animations\n",
		},
		{
			name: "search for extensions with qualifier flags",
			args: []string{"search", "--license", "GPLv3", "--owner", "jillvalentine"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{}
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
				}
			},
			httpStubs: func(reg *httpmock.Registry) {
				values := url.Values{
					"page":     []string{"1"},
					"per_page": []string{"30"},
					"q":        []string{"license:GPLv3 topic:gh-extension user:jillvalentine"},
				}
				results := searchResults(1)
				reg.Register(
					httpmock.QueryMatcher("GET", literal_0923, values),
					httpmock.JSONResponse(results),
				)
			},
			wantStdout: "\tvilmibm/gh-screensaver\tterminal animations\n",
		},
		{
			name: "search for extensions with web mode",
			args: []string{"search", "--web"},
			browseStubs: func(b *browser.Stub) func(*testing.T) {
				return func(t *testing.T) {
					b.Verify(t, "https://github.com/search?q=topic%3Agh-extension&type=repositories")
				}
			},
		},
		{
			name: "install an extension",
			args: []string{"install", "owner/gh-some-ext"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{}
				}
				em.InstallFunc = func(_ ghrepo.Interface, _ string) error {
					return nil
				}
				return func(t *testing.T) {
					installCalls := em.InstallCalls()
					assert.Equal(t, 1, len(installCalls))
					assert.Equal(t, "gh-some-ext", installCalls[0].InterfaceMoqParam.RepoName())
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
				}
			},
		},
		{
			name: "install an extension with same name as existing extension",
			args: []string{"install", literal_7841},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					e := &Extension{path: "owner2/gh-existing-ext", owner: "owner2"}
					return []extensions.Extension{e}
				}
				return func(t *testing.T) {
					calls := em.ListCalls()
					assert.Equal(t, 1, len(calls))
				}
			},
			wantErr: true,
			errMsg:  "there is already an installed extension that provides the \"existing-ext\" command",
		},
		{
			name: "install an already installed extension",
			args: []string{"install", literal_7841},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					e := &Extension{path: literal_7841, owner: "owner"}
					return []extensions.Extension{e}
				}
				return func(t *testing.T) {
					calls := em.ListCalls()
					assert.Equal(t, 1, len(calls))
				}
			},
			wantStderr: "! Extension owner/gh-existing-ext is already installed\n",
		},
		{
			name: "install local extension",
			args: []string{"install", "."},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.InstallLocalFunc = func(dir string) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.InstallLocalCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, tempDir, normalizeDir(calls[0].Dir))
				}
			},
		},
		{
			name: "error extension not found",
			args: []string{"install", "owner/gh-some-ext"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{}
				}
				em.InstallFunc = func(_ ghrepo.Interface, _ string) error {
					return repositoryNotFoundErr
				}
				return func(t *testing.T) {
					installCalls := em.InstallCalls()
					assert.Equal(t, 1, len(installCalls))
					assert.Equal(t, "gh-some-ext", installCalls[0].InterfaceMoqParam.RepoName())
				}
			},
			wantErr: true,
			errMsg:  "X Could not find extension 'owner/gh-some-ext' on host github.com",
		},
		{
			name:    "install local extension with pin",
			args:    []string{"install", ".", "--pin", "v1.0.0"},
			wantErr: true,
			errMsg:  "local extensions cannot be pinned",
			isTTY:   true,
		},
		{
			name:    "upgrade argument error",
			args:    []string{"upgrade"},
			wantErr: true,
			errMsg:  "specify an extension to upgrade or `--all`",
		},
		{
			name:    "upgrade --all with extension name error",
			args:    []string{"upgrade", "test", "--all"},
			wantErr: true,
			errMsg:  "cannot use `--all` with extension name",
		},
		{
			name: "upgrade an extension",
			args: []string{"upgrade", "hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: literal_3205,
		},
		{
			name: "upgrade an extension dry run",
			args: []string{"upgrade", "hello", "--dry-run"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.EnableDryRunModeFunc = func() {}
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					dryRunCalls := em.EnableDryRunModeCalls()
					assert.Equal(t, 1, len(dryRunCalls))
					upgradeCalls := em.UpgradeCalls()
					assert.Equal(t, 1, len(upgradeCalls))
					assert.Equal(t, "hello", upgradeCalls[0].Name)
					assert.False(t, upgradeCalls[0].Force)
				}
			},
			isTTY:      true,
			wantStdout: "✓ Would have upgraded extension\n",
		},
		{
			name: "upgrade an extension notty",
			args: []string{"upgrade", "hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY: false,
		},
		{
			name: "upgrade an up-to-date extension",
			args: []string{"upgrade", "hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					// An already up to date extension returns the same response
					// as an one that has been upgraded.
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: literal_3205,
		},
		{
			name: "upgrade extension error",
			args: []string{"upgrade", "hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return errors.New("oh no")
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      false,
			wantErr:    true,
			errMsg:     "SilentError",
			wantStdout: "",
			wantStderr: "X Failed upgrading extension hello: oh no\n",
		},
		{
			name: "upgrade an extension gh-prefix",
			args: []string{"upgrade", literal_5943},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: literal_3205,
		},
		{
			name: "upgrade an extension full name",
			args: []string{"upgrade", "monalisa/gh-hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: literal_3205,
		},
		{
			name: "upgrade all",
			args: []string{"upgrade", "--all"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "", calls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: "✓ Successfully upgraded extensions\n",
		},
		{
			name: "upgrade all dry run",
			args: []string{"upgrade", "--all", "--dry-run"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.EnableDryRunModeFunc = func() {}
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					dryRunCalls := em.EnableDryRunModeCalls()
					assert.Equal(t, 1, len(dryRunCalls))
					upgradeCalls := em.UpgradeCalls()
					assert.Equal(t, 1, len(upgradeCalls))
					assert.Equal(t, "", upgradeCalls[0].Name)
					assert.False(t, upgradeCalls[0].Force)
				}
			},
			isTTY:      true,
			wantStdout: "✓ Would have upgraded extensions\n",
		},
		{
			name: "upgrade all none installed",
			args: []string{"upgrade", "--all"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return noExtensionsInstalledError
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "", calls[0].Name)
				}
			},
			isTTY:   true,
			wantErr: true,
			errMsg:  "no installed extensions found",
		},
		{
			name: "upgrade all notty",
			args: []string{"upgrade", "--all"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.UpgradeCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "", calls[0].Name)
				}
			},
			isTTY: false,
		},
		{
			name: "remove extension tty",
			args: []string{"remove", "hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.RemoveFunc = func(name string) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.RemoveCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: "✓ Removed extension hello\n",
		},
		{
			name: "remove extension nontty",
			args: []string{"remove", "hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.RemoveFunc = func(name string) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.RemoveCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      false,
			wantStdout: "",
		},
		{
			name: "remove extension gh-prefix",
			args: []string{"remove", literal_5943},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.RemoveFunc = func(name string) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.RemoveCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      false,
			wantStdout: "",
		},
		{
			name: "remove extension full name",
			args: []string{"remove", "monalisa/gh-hello"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.RemoveFunc = func(name string) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.RemoveCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, "hello", calls[0].Name)
				}
			},
			isTTY:      false,
			wantStdout: "",
		},
		{
			name: "list extensions",
			args: []string{"list"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					ex1 := &Extension{path: "cli/gh-test", url: "https://github.com/cli/gh-test", currentVersion: "1"}
					ex2 := &Extension{path: "cli/gh-test2", url: "https://github.com/cli/gh-test2", currentVersion: "1"}
					return []extensions.Extension{ex1, ex2}
				}
				return func(t *testing.T) {
					calls := em.ListCalls()
					assert.Equal(t, 1, len(calls))
				}
			},
			wantStdout: "gh test\tcli/gh-test\t1\ngh test2\tcli/gh-test2\t1\n",
		},
		{
			name: "create extension interactive",
			args: []string{"create"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.CreateFunc = func(name string, tmplType extensions.ExtTemplateType) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.CreateCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, literal_3462, calls[0].Name)
				}
			},
			isTTY: true,
			prompterStubs: func(pm *prompter.PrompterMock) {
				pm.InputFunc = func(prompt, defVal string) (string, error) {
					if prompt == "Extension name:" {
						return "test", nil
					}
					return "", nil
				}
				pm.SelectFunc = func(prompt, defVal string, opts []string) (int, error) {
					return prompter.IndexFor(opts, "Script (Bash, Ruby, Python, etc)")
				}
			},
			wantStdout: heredoc.Doc(`
				✓ Created directory gh-test
				✓ Initialized git repository
				✓ Made initial commit
				✓ Set up extension scaffolding

				gh-test is ready for development!

				Next Steps
				- run 'cd gh-test; gh extension install .; gh test' to see your new extension in action
				- run 'gh repo create' to share your extension with others

				For more information on writing extensions:
				https://docs.github.com/github-cli/github-cli/creating-github-cli-extensions
			`),
		},
		{
			name: "create extension with arg, --precompiled=go",
			args: []string{"create", "test", "--precompiled", "go"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.CreateFunc = func(name string, tmplType extensions.ExtTemplateType) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.CreateCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, literal_3462, calls[0].Name)
				}
			},
			isTTY: true,
			wantStdout: heredoc.Doc(`
				✓ Created directory gh-test
				✓ Initialized git repository
				✓ Made initial commit
				✓ Set up extension scaffolding
				✓ Downloaded Go dependencies
				✓ Built gh-test binary

				gh-test is ready for development!

				Next Steps
				- run 'cd gh-test; gh extension install .; gh test' to see your new extension in action
				- run 'go build && gh test' to see changes in your code as you develop
				- run 'gh repo create' to share your extension with others

				For more information on writing extensions:
				https://docs.github.com/github-cli/github-cli/creating-github-cli-extensions
			`),
		},
		{
			name: "create extension with arg, --precompiled=other",
			args: []string{"create", "test", "--precompiled", "other"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.CreateFunc = func(name string, tmplType extensions.ExtTemplateType) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.CreateCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, literal_3462, calls[0].Name)
				}
			},
			isTTY: true,
			wantStdout: heredoc.Doc(`
				✓ Created directory gh-test
				✓ Initialized git repository
				✓ Made initial commit
				✓ Set up extension scaffolding

				gh-test is ready for development!

				Next Steps
				- run 'cd gh-test; gh extension install .' to install your extension locally
				- fill in script/build.sh with your compilation script for automated builds
				- compile a gh-test binary locally and run 'gh test' to see changes
				- run 'gh repo create' to share your extension with others

				For more information on writing extensions:
				https://docs.github.com/github-cli/github-cli/creating-github-cli-extensions
			`),
		},
		{
			name: "create extension tty with argument",
			args: []string{"create", "test"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.CreateFunc = func(name string, tmplType extensions.ExtTemplateType) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.CreateCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, literal_3462, calls[0].Name)
				}
			},
			isTTY: true,
			wantStdout: heredoc.Doc(`
				✓ Created directory gh-test
				✓ Initialized git repository
				✓ Made initial commit
				✓ Set up extension scaffolding

				gh-test is ready for development!

				Next Steps
				- run 'cd gh-test; gh extension install .; gh test' to see your new extension in action
				- run 'gh repo create' to share your extension with others

				For more information on writing extensions:
				https://docs.github.com/github-cli/github-cli/creating-github-cli-extensions
			`),
		},
		{
			name: "create extension tty with argument commit fails",
			args: []string{"create", "test"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.CreateFunc = func(name string, tmplType extensions.ExtTemplateType) error {
					return ErrInitialCommitFailed
				}
				return func(t *testing.T) {
					calls := em.CreateCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, literal_3462, calls[0].Name)
				}
			},
			isTTY: true,
			wantStdout: heredoc.Doc(`
				✓ Created directory gh-test
				✓ Initialized git repository
				X Made initial commit
				✓ Set up extension scaffolding

				gh-test is ready for development!

				Next Steps
				- run 'cd gh-test; gh extension install .; gh test' to see your new extension in action
				- run 'gh repo create' to share your extension with others

				For more information on writing extensions:
				https://docs.github.com/github-cli/github-cli/creating-github-cli-extensions
			`),
		},
		{
			name: "create extension notty",
			args: []string{"create", literal_3462},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.CreateFunc = func(name string, tmplType extensions.ExtTemplateType) error {
					return nil
				}
				return func(t *testing.T) {
					calls := em.CreateCalls()
					assert.Equal(t, 1, len(calls))
					assert.Equal(t, literal_3462, calls[0].Name)
				}
			},
			isTTY:      false,
			wantStdout: "",
		},
		{
			name: "exec extension missing",
			args: []string{"exec", "invalid"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.DispatchFunc = func(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
					return false, nil
				}
				return func(t *testing.T) {
					calls := em.DispatchCalls()
					assert.Equal(t, 1, len(calls))
					assert.EqualValues(t, []string{"invalid"}, calls[0].Args)
				}
			},
			wantErr: true,
			errMsg:  `extension "invalid" not found`,
		},
		{
			name: "exec extension with arguments",
			args: []string{"exec", "test", "arg1", "arg2", "--flag1"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.DispatchFunc = func(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
					fmt.Fprintf(stdout, "test output")
					return true, nil
				}
				return func(t *testing.T) {
					calls := em.DispatchCalls()
					assert.Equal(t, 1, len(calls))
					assert.EqualValues(t, []string{"test", "arg1", "arg2", "--flag1"}, calls[0].Args)
				}
			},
			wantStdout: "test output",
		},
		{
			name:    "browse",
			args:    []string{"browse"},
			wantErr: true,
			errMsg:  "this command runs an interactive UI and needs to be run in a terminal",
		},
		{
			name: "force install when absent",
			args: []string{"install", literal_5312, "--force"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{}
				}
				em.InstallFunc = func(_ ghrepo.Interface, _ string) error {
					return nil
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
					installCalls := em.InstallCalls()
					assert.Equal(t, 1, len(installCalls))
					assert.Equal(t, literal_5943, installCalls[0].InterfaceMoqParam.RepoName())
				}
			},
			isTTY:      true,
			wantStdout: "✓ Installed extension owner/gh-hello\n",
		},
		{
			name: "force install when present",
			args: []string{"install", literal_5312, "--force"},
			managerStubs: func(em *extensions.ExtensionManagerMock) func(*testing.T) {
				em.ListFunc = func() []extensions.Extension {
					return []extensions.Extension{
						&Extension{path: literal_5312, owner: "owner"},
					}
				}
				em.InstallFunc = func(_ ghrepo.Interface, _ string) error {
					return nil
				}
				em.UpgradeFunc = func(name string, force bool) error {
					return nil
				}
				return func(t *testing.T) {
					listCalls := em.ListCalls()
					assert.Equal(t, 1, len(listCalls))
					installCalls := em.InstallCalls()
					assert.Equal(t, 0, len(installCalls))
					upgradeCalls := em.UpgradeCalls()
					assert.Equal(t, 1, len(upgradeCalls))
					assert.Equal(t, "hello", upgradeCalls[0].Name)
				}
			},
			isTTY:      true,
			wantStdout: literal_3205,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			var assertFunc func(*testing.T)
			em := &extensions.ExtensionManagerMock{}
			if tt.managerStubs != nil {
				assertFunc = tt.managerStubs(em)
			}

			pm := &prompter.PrompterMock{}
			if tt.prompterStubs != nil {
				tt.prompterStubs(pm)
			}

			reg := httpmock.Registry{}
			defer reg.Verify(t)
			client := http.Client{Transport: &reg}

			if tt.httpStubs != nil {
				tt.httpStubs(&reg)
			}

			var assertBrowserFunc func(*testing.T)
			browseStub := &browser.Stub{}
			if tt.browseStubs != nil {
				assertBrowserFunc = tt.browseStubs(browseStub)
			}

			f := cmdutil.Factory{
				Config: func() (config.Config, error) {
					return config.NewBlankConfig(), nil
				},
				IOStreams:        ios,
				ExtensionManager: em,
				Prompter:         pm,
				Browser:          browseStub,
				HttpClient: func() (*http.Client, error) {
					return &client, nil
				},
			}

			cmd := NewCmdExtension(&f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err := cmd.ExecuteC()
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			if assertFunc != nil {
				assertFunc(t)
			}

			if assertBrowserFunc != nil {
				assertBrowserFunc(t)
			}

			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())
		})
	}
}

func normalizeDir(d string) string {
	return strings.TrimPrefix(d, "/private")
}

func TestCheckValidExtension(t *testing.T) {
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(&cobra.Command{Use: "help"})
	rootCmd.AddCommand(&cobra.Command{Use: "auth"})

	m := &extensions.ExtensionManagerMock{
		ListFunc: func() []extensions.Extension {
			return []extensions.Extension{
				&extensions.ExtensionMock{
					OwnerFunc: func() string { return "monalisa" },
					NameFunc:  func() string { return "screensaver" },
				},
				&extensions.ExtensionMock{
					OwnerFunc: func() string { return "monalisa" },
					NameFunc:  func() string { return "triage" },
				},
			}
		},
	}

	type args struct {
		rootCmd  *cobra.Command
		manager  extensions.ExtensionManager
		extName  string
		extOwner string
	}
	tests := []struct {
		name      string
		args      args
		wantError string
	}{
		{
			name: "valid extension",
			args: args{
				rootCmd:  rootCmd,
				manager:  m,
				extOwner: "monalisa",
				extName:  literal_5943,
			},
		},
		{
			name: "invalid extension name",
			args: args{
				rootCmd:  rootCmd,
				manager:  m,
				extOwner: "monalisa",
				extName:  "gherkins",
			},
			wantError: "extension repository name must start with `gh-`",
		},
		{
			name: "clashes with built-in command",
			args: args{
				rootCmd:  rootCmd,
				manager:  m,
				extOwner: "monalisa",
				extName:  "gh-auth",
			},
			wantError: "\"auth\" matches the name of a built-in command or alias",
		},
		{
			name: "clashes with an installed extension",
			args: args{
				rootCmd:  rootCmd,
				manager:  m,
				extOwner: "cli",
				extName:  literal_2083,
			},
			wantError: "there is already an installed extension that provides the \"triage\" command",
		},
		{
			name: "clashes with same extension",
			args: args{
				rootCmd:  rootCmd,
				manager:  m,
				extOwner: "monalisa",
				extName:  literal_2083,
			},
			wantError: "alreadyInstalledError",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := checkValidExtension(tt.args.rootCmd, tt.args.manager, tt.args.extName, tt.args.extOwner)
			if tt.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantError)
			}
		})
	}
}

func searchResults(numResults int) interface{} {
	result := map[string]interface{}{
		"incomplete_results": false,
		"total_count":        4,
		"items": []interface{}{
			map[string]interface{}{
				"name":        "gh-screensaver",
				"full_name":   "vilmibm/gh-screensaver",
				"description": "terminal animations",
				"owner": map[string]interface{}{
					"login": "vilmibm",
				},
			},
			map[string]interface{}{
				"name":        "gh-cool",
				"full_name":   "cli/gh-cool",
				"description": "it's just cool ok",
				"owner": map[string]interface{}{
					"login": "cli",
				},
			},
			map[string]interface{}{
				"name":        literal_2083,
				"full_name":   "samcoe/gh-triage",
				"description": "helps with triage",
				"owner": map[string]interface{}{
					"login": "samcoe",
				},
			},
			map[string]interface{}{
				"name":        "gh-gei",
				"full_name":   "github/gh-gei",
				"description": "something something enterprise",
				"owner": map[string]interface{}{
					"login": "github",
				},
			},
		},
	}
	if len(result["items"].([]interface{})) > numResults {
		fewerItems := result["items"].([]interface{})[0:numResults]
		result["items"] = fewerItems
	}
	return result
}

const literal_5902 = "https://github.com/vilmibm/gh-screensaver"

const literal_4230 = "https://github.com/github/gh-gei"

const literal_9853 = "topic:gh-extension"

const literal_0923 = "search/repositories"

const literal_7841 = "owner/gh-existing-ext"

const literal_3205 = "✓ Successfully upgraded extension\n"

const literal_5943 = "gh-hello"

const literal_3462 = "gh-test"

const literal_5312 = "owner/gh-hello"

const literal_2083 = "gh-triage"
