package download

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/pkg/cmd/release/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdDownload(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		isTTY   bool
		want    DownloadOptions
		wantErr string
	}{
		{
			name:  "version argument",
			args:  literal_7048,
			isTTY: true,
			want: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string(nil),
				Destination:  ".",
				Concurrency:  5,
			},
		},
		{
			name:  "version and file pattern",
			args:  "v1.2.3 -p *.tgz",
			isTTY: true,
			want: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string{"*.tgz"},
				Destination:  ".",
				Concurrency:  5,
			},
		},
		{
			name:  "multiple file patterns",
			args:  "v1.2.3 -p 1 -p 2,3",
			isTTY: true,
			want: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string{"1", "2,3"},
				Destination:  ".",
				Concurrency:  5,
			},
		},
		{
			name:  "version and destination",
			args:  "v1.2.3 -D tmp/assets",
			isTTY: true,
			want: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string(nil),
				Destination:  "tmp/assets",
				Concurrency:  5,
			},
		},
		{
			name:  "download latest",
			args:  "-p *",
			isTTY: true,
			want: DownloadOptions{
				TagName:      "",
				FilePatterns: []string{"*"},
				Destination:  ".",
				Concurrency:  5,
			},
		},
		{
			name:  "download archive with valid option",
			args:  "v1.2.3 -A zip",
			isTTY: true,
			want: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string(nil),
				Destination:  ".",
				ArchiveType:  "zip",
				Concurrency:  5,
			},
		},
		{
			name:  "download to output with valid option",
			args:  "v1.2.3 -A zip -O ./sample.zip",
			isTTY: true,
			want: DownloadOptions{
				OutputFile:   "./sample.zip",
				TagName:      literal_7048,
				FilePatterns: []string(nil),
				Destination:  ".",
				ArchiveType:  "zip",
				Concurrency:  5,
			},
		},
		{
			name:    "no arguments",
			args:    "",
			isTTY:   true,
			wantErr: "`--pattern` or `--archive` is required when downloading the latest release",
		},
		{
			name:    "simultaneous pattern and archive arguments",
			args:    "-p * -A zip",
			isTTY:   true,
			wantErr: "specify only one of '--pattern' or '--archive'",
		},
		{
			name:    "invalid archive argument",
			args:    "v1.2.3 -A abc",
			isTTY:   true,
			wantErr: "the value for `--archive` must be one of \"zip\" or \"tar.gz\"",
		},
		{
			name:    "simultaneous output and destination flags",
			args:    "v1.2.3 -O ./file.xyz -D ./destination",
			isTTY:   true,
			wantErr: "specify only one of `--dir` or `--output`",
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

			var opts *DownloadOptions
			cmd := NewCmdDownload(f, func(o *DownloadOptions) error {
				opts = o
				return nil
			})
			cmd.PersistentFlags().StringP("repo", "R", "", "")

			argv, err := shlex.Split(tt.args)
			assert.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.want.TagName, opts.TagName)
			assert.Equal(t, tt.want.FilePatterns, opts.FilePatterns)
			assert.Equal(t, tt.want.Destination, opts.Destination)
			assert.Equal(t, tt.want.Concurrency, opts.Concurrency)
			assert.Equal(t, tt.want.OutputFile, opts.OutputFile)
		})
	}
}

func TestDownloadRun(t *testing.T) {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not determine working directory: %v", err)
	}

	tests := []struct {
		name       string
		isTTY      bool
		opts       DownloadOptions
		wantErr    string
		wantStdout string
		wantStderr string
		wantFiles  []string
	}{
		{
			name:  "download all assets",
			isTTY: true,
			opts: DownloadOptions{
				TagName:     literal_7048,
				Destination: ".",
				Concurrency: 2,
			},
			wantStdout: ``,
			wantStderr: ``,
			wantFiles: []string{
				"linux.tgz",
				"windows-32bit.zip",
				literal_7150,
			},
		},
		{
			name:  "download assets matching pattern into destination directory",
			isTTY: true,
			opts: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string{"windows-*.zip"},
				Destination:  "tmp/assets",
				Concurrency:  2,
			},
			wantStdout: ``,
			wantStderr: ``,
			wantFiles: []string{
				"tmp/assets/windows-32bit.zip",
				"tmp/assets/windows-64bit.zip",
			},
		},
		{
			name:  "no match for pattern",
			isTTY: true,
			opts: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string{"linux*.zip"},
				Destination:  ".",
				Concurrency:  2,
			},
			wantStdout: ``,
			wantStderr: ``,
			wantErr:    "no assets match the file pattern",
		},
		{
			name:  "download archive in zip format into destination directory",
			isTTY: true,
			opts: DownloadOptions{
				TagName:     literal_7048,
				ArchiveType: "zip",
				Destination: literal_4236,
				Concurrency: 2,
			},
			wantStdout: ``,
			wantStderr: ``,
			wantFiles: []string{
				"tmp/packages/zipball.zip",
			},
		},
		{
			name:  "download archive in `tar.gz` format into destination directory",
			isTTY: true,
			opts: DownloadOptions{
				TagName:     literal_7048,
				ArchiveType: "tar.gz",
				Destination: literal_4236,
				Concurrency: 2,
			},
			wantStdout: ``,
			wantStderr: ``,
			wantFiles: []string{
				"tmp/packages/tarball.tgz",
			},
		},
		{
			name:  "download archive in `tar.gz` format into output option",
			isTTY: true,
			opts: DownloadOptions{
				OutputFile:  "./tmp/my-tarball.tgz",
				TagName:     literal_7048,
				Destination: "",
				Concurrency: 2,
				ArchiveType: "tar.gz",
			},
			wantStdout: ``,
			wantStderr: ``,
			wantFiles: []string{
				"tmp/my-tarball.tgz",
			},
		},
		{
			name:  "download single asset from matching pattern into output option",
			isTTY: true,
			opts: DownloadOptions{
				OutputFile:   "./tmp/my-tarball.tgz",
				TagName:      literal_7048,
				Destination:  "",
				Concurrency:  2,
				FilePatterns: []string{"*windows-32bit.zip"},
			},
			wantStdout: ``,
			wantStderr: ``,
			wantFiles: []string{
				"tmp/my-tarball.tgz",
			},
		},
		{
			name:  "download single asset from matching pattern into output 'stdout´",
			isTTY: true,
			opts: DownloadOptions{
				OutputFile:   "-",
				TagName:      literal_7048,
				Destination:  "",
				Concurrency:  2,
				FilePatterns: []string{"*windows-32bit.zip"},
			},
			wantStdout: `1234`,
			wantStderr: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := os.Chdir(tempDir); err == nil {
				t.Cleanup(func() { _ = os.Chdir(oldwd) })
			} else {
				t.Fatal(err)
			}

			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			fakeHTTP := &httpmock.Registry{}
			shared.StubFetchRelease(t, fakeHTTP, "OWNER", "REPO", tt.opts.TagName, `{
				"assets": [
					{ "name": "windows-32bit.zip", "size": 12,
					  "url": "https://api.github.com/assets/1234" },
					{ "name": literal_7150, "size": 34,
					  "url": "https://api.github.com/assets/3456" },
					{ "name": "linux.tgz", "size": 56,
					  "url": "https://api.github.com/assets/5678" }
				],
				"tarball_url": "https://api.github.com/repos/OWNER/REPO/tarball/v1.2.3",
				"zipball_url": "https://api.github.com/repos/OWNER/REPO/zipball/v1.2.3"
			}`)
			fakeHTTP.Register(httpmock.REST("GET", "assets/1234"), httpmock.StringResponse(`1234`))
			fakeHTTP.Register(httpmock.REST("GET", "assets/3456"), httpmock.StringResponse(`3456`))
			fakeHTTP.Register(httpmock.REST("GET", "assets/5678"), httpmock.StringResponse(`5678`))

			fakeHTTP.Register(
				httpmock.REST(
					"GET",
					"repos/OWNER/REPO/tarball/v1.2.3",
				),
				httpmock.WithHeader(
					httpmock.StringResponse("somedata"), literal_7162, "attachment; filename=tarball.tgz",
				),
			)

			fakeHTTP.Register(
				httpmock.REST(
					"GET",
					literal_4563,
				),
				httpmock.WithHeader(
					httpmock.StringResponse("somedata"), literal_7162, literal_8293,
				),
			)

			tt.opts.IO = ios
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: fakeHTTP}, nil
			}
			tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.FromFullName(literal_2741)
			}

			err := downloadRun(&tt.opts)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)

			var expectedAcceptHeader = "application/octet-stream"
			if len(tt.opts.ArchiveType) > 0 {
				expectedAcceptHeader = "application/octet-stream, application/json"
			}

			for _, req := range fakeHTTP.Requests {
				if req.Method != "GET" || req.URL.Path != "repos/OWNER/REPO/releases/tags/v1.2.3" {
					// skip non-asset download requests
					continue
				}
				assert.Equal(t, expectedAcceptHeader, req.Header.Get("Accept"), "for request %s", req.URL)
			}

			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantStderr, stderr.String())

			downloadedFiles, err := listFiles(".")
			assert.NoError(t, err)
			assert.Equal(t, tt.wantFiles, downloadedFiles)
		})
	}
}

func TestDownloadRun_cloberAndSkip(t *testing.T) {
	oldAssetContents := "older copy to be clobbered"
	oldZipballContents := "older zipball to be clobbered"
	// this should be shorter than oldAssetContents and oldZipballContents
	newContents := "somedata"

	tests := []struct {
		name            string
		opts            DownloadOptions
		httpStubs       func(*httpmock.Registry)
		wantErr         string
		wantFileSize    int64
		wantArchiveSize int64
	}{
		{
			name: "no clobber or skip",
			opts: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string{literal_7150},
				Destination:  literal_4236,
				Concurrency:  2,
			},
			wantErr:         "already exists (use `--clobber` to overwrite file or `--skip-existing` to skip file)",
			wantFileSize:    int64(len(oldAssetContents)),
			wantArchiveSize: int64(len(oldZipballContents)),
		},
		{
			name: "clobber",
			opts: DownloadOptions{
				TagName:           literal_7048,
				FilePatterns:      []string{literal_7150},
				Destination:       literal_4236,
				Concurrency:       2,
				OverwriteExisting: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(httpmock.REST("GET", "assets/3456"), httpmock.StringResponse(newContents))
			},
			wantFileSize:    int64(len(newContents)),
			wantArchiveSize: int64(len(oldZipballContents)),
		},
		{
			name: "clobber archive",
			opts: DownloadOptions{
				TagName:           literal_7048,
				ArchiveType:       "zip",
				Destination:       literal_4236,
				Concurrency:       2,
				OverwriteExisting: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_4563),
					httpmock.WithHeader(
						httpmock.StringResponse(newContents), literal_7162, literal_8293,
					),
				)
			},
			wantFileSize:    int64(len(oldAssetContents)),
			wantArchiveSize: int64(len(newContents)),
		},
		{
			name: "skip",
			opts: DownloadOptions{
				TagName:      literal_7048,
				FilePatterns: []string{literal_7150},
				Destination:  literal_4236,
				Concurrency:  2,
				SkipExisting: true,
			},
			wantFileSize:    int64(len(oldAssetContents)),
			wantArchiveSize: int64(len(oldZipballContents)),
		},
		{
			name: "skip archive",
			opts: DownloadOptions{
				TagName:      literal_7048,
				ArchiveType:  "zip",
				Destination:  literal_4236,
				Concurrency:  2,
				SkipExisting: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_4563),
					httpmock.WithHeader(
						httpmock.StringResponse(newContents), literal_7162, literal_8293,
					),
				)
			},
			wantFileSize:    int64(len(oldAssetContents)),
			wantArchiveSize: int64(len(oldZipballContents)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			dest := filepath.Join(tempDir, tt.opts.Destination)
			err := os.MkdirAll(dest, 0755)
			assert.NoError(t, err)
			file := filepath.Join(dest, literal_7150)
			archive := filepath.Join(dest, "zipball.zip")
			f1, err := os.Create(file)
			assert.NoError(t, err)
			_, err = f1.WriteString(oldAssetContents)
			assert.NoError(t, err)
			f1.Close()
			f2, err := os.Create(archive)
			assert.NoError(t, err)
			_, err = f2.WriteString(oldZipballContents)
			assert.NoError(t, err)
			f2.Close()

			tt.opts.Destination = dest

			ios, _, _, _ := iostreams.Test()
			tt.opts.IO = ios

			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			shared.StubFetchRelease(t, reg, "OWNER", "REPO", literal_7048, `{
				"assets": [
					{ "name": literal_7150, "size": 34,
					  "url": "https://api.github.com/assets/3456" }
				],
				"tarball_url": "https://api.github.com/repos/OWNER/REPO/tarball/v1.2.3",
				"zipball_url": "https://api.github.com/repos/OWNER/REPO/zipball/v1.2.3"
			}`)
			if tt.httpStubs != nil {
				tt.httpStubs(reg)
			}
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}

			tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
				return ghrepo.FromFullName(literal_2741)
			}

			err = downloadRun(&tt.opts)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			fs, err := os.Stat(file)
			assert.NoError(t, err)
			as, err := os.Stat(archive)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantFileSize, fs.Size())
			assert.Equal(t, tt.wantArchiveSize, as.Size())
		})
	}
}

func TestDownloadRun_windowsReservedFilename(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}

	tagName := literal_7048

	ios, _, _, _ := iostreams.Test()

	reg := &httpmock.Registry{}
	defer reg.Verify(t)

	shared.StubFetchRelease(t, reg, "OWNER", "REPO", tagName, `{
		"assets": [
			{ "name": "valid-asset.zip", "size": 12,
			  "url": "https://api.github.com/assets/1234" },
			{ "name": "valid-asset-2.zip", "size": 34,
			  "url": "https://api.github.com/assets/3456" },
			{ "name": "CON.tgz", "size": 56,
			  "url": "https://api.github.com/assets/5678" }
		],
		"tarball_url": "https://api.github.com/repos/OWNER/REPO/tarball/v1.2.3",
		"zipball_url": "https://api.github.com/repos/OWNER/REPO/zipball/v1.2.3"
	}`)

	opts := &DownloadOptions{
		IO: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.FromFullName(literal_2741)
		},
		TagName: tagName,
	}

	err := downloadRun(opts)

	assert.EqualError(t, err, `unable to download release due to asset with reserved filename "CON.tgz"`)
}

func TestIsWindowsReservedFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "non-reserved filename",
			filename: "test",
			want:     false,
		},
		{
			name:     "non-reserved filename with file type extension",
			filename: "test.tar.gz",
			want:     false,
		},
		{
			name:     "reserved filename",
			filename: "NUL",
			want:     true,
		},
		{
			name:     "reserved filename with file type extension",
			filename: "NUL.tar.gz",
			want:     true,
		},
		{
			name:     "reserved filename with mixed type case",
			filename: "NuL",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isWindowsReservedFilename(tt.filename))
		})
	}
}

func listFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(p string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			rp, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rp))
		}
		return err
	})
	return files, err
}

const literal_7048 = "v1.2.3"

const literal_7150 = "windows-64bit.zip"

const literal_4236 = "tmp/packages"

const literal_7162 = "content-disposition"

const literal_4563 = "repos/OWNER/REPO/zipball/v1.2.3"

const literal_8293 = "attachment; filename=zipball.zip"

const literal_2741 = "OWNER/REPO"
