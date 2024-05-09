package githubtemplate

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func TestFindNonLegacy(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", literal_9536)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		rootDir string
		name    string
	}
	tests := []struct {
		name    string
		prepare []string
		args    args
		want    []string
	}{
		{
			name: "Legacy templates ignored",
			prepare: []string{
				literal_4195,
				"ISSUE_TEMPLATE",
				literal_7506,
				"issue_template.txt",
				"pull_request_template.md",
				literal_4830,
				literal_8065,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: []string{},
		},
		{
			name: "Template folder in .github takes precedence",
			prepare: []string{
				literal_3056,
				literal_0634,
				"ISSUE_TEMPLATE/abc.md",
				".github/ISSUE_TEMPLATE/abc.md",
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: []string{
				path.Join(tmpdir, ".github/ISSUE_TEMPLATE/abc.md"),
			},
		},
		{
			name: "Template folder in root",
			prepare: []string{
				literal_3056,
				literal_0634,
				"ISSUE_TEMPLATE/abc.md",
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: []string{
				path.Join(tmpdir, "ISSUE_TEMPLATE/abc.md"),
			},
		},
		{
			name: "Template folder in docs",
			prepare: []string{
				literal_3056,
				literal_0634,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: []string{
				path.Join(tmpdir, literal_0634),
			},
		},
		{
			name: "Multiple templates in template folder",
			prepare: []string{
				".github/ISSUE_TEMPLATE/nope.md",
				literal_4710,
				".github/PULL_REQUEST_TEMPLATE/one.md",
				".github/PULL_REQUEST_TEMPLATE/two.md",
				".github/PULL_REQUEST_TEMPLATE/three.md",
				"docs/pull_request_template.md",
			},
			args: args{
				rootDir: tmpdir,
				name:    "PuLl_ReQuEsT_TeMpLaTe",
			},
			want: []string{
				path.Join(tmpdir, ".github/PULL_REQUEST_TEMPLATE/one.md"),
				path.Join(tmpdir, ".github/PULL_REQUEST_TEMPLATE/three.md"),
				path.Join(tmpdir, ".github/PULL_REQUEST_TEMPLATE/two.md"),
			},
		},
		{
			name: "Empty template directories",
			prepare: []string{
				".github/ISSUE_TEMPLATE/.keep",
				".docs/ISSUE_TEMPLATE/.keep",
				"ISSUE_TEMPLATE/.keep",
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, p := range tt.prepare {
				fp := path.Join(tmpdir, p)
				_ = os.MkdirAll(path.Dir(fp), 0700)
				file, err := os.Create(fp)
				if err != nil {
					t.Fatal(err)
				}
				file.Close()
			}

			if got := FindNonLegacy(tt.args.rootDir, tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
		os.RemoveAll(tmpdir)
	}
}

func TestFindLegacy(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", literal_9536)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		rootDir string
		name    string
	}
	tests := []struct {
		name    string
		prepare []string
		args    args
		want    string
	}{
		{
			name: "Template in root",
			prepare: []string{
				literal_4195,
				literal_7506,
				"issue_template.txt",
				"pull_request_template.md",
				literal_8065,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: path.Join(tmpdir, literal_7506),
		},
		{
			name: "No extension",
			prepare: []string{
				literal_4195,
				"issue_template",
				literal_8065,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: path.Join(tmpdir, "issue_template"),
		},
		{
			name: "Dash instead of underscore",
			prepare: []string{
				literal_4195,
				"issue-template.txt",
				literal_8065,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: path.Join(tmpdir, "issue-template.txt"),
		},
		{
			name: "Template in .github takes precedence",
			prepare: []string{
				literal_3056,
				literal_4830,
				literal_8065,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: path.Join(tmpdir, literal_4830),
		},
		{
			name: "Template in docs",
			prepare: []string{
				literal_4195,
				literal_8065,
			},
			args: args{
				rootDir: tmpdir,
				name:    "ISSUE_TEMPLATE",
			},
			want: path.Join(tmpdir, literal_8065),
		},
		{
			name: "Non legacy templates ignored",
			prepare: []string{
				".github/PULL_REQUEST_TEMPLATE/abc.md",
				"PULL_REQUEST_TEMPLATE/abc.md",
				"docs/PULL_REQUEST_TEMPLATE/abc.md",
				literal_4710,
			},
			args: args{
				rootDir: tmpdir,
				name:    "PuLl_ReQuEsT_TeMpLaTe",
			},
			want: path.Join(tmpdir, literal_4710),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, p := range tt.prepare {
				fp := path.Join(tmpdir, p)
				_ = os.MkdirAll(path.Dir(fp), 0700)
				file, err := os.Create(fp)
				if err != nil {
					t.Fatal(err)
				}
				file.Close()
			}

			got := FindLegacy(tt.args.rootDir, tt.args.name)
			if got == "" {
				t.Errorf("FindLegacy() = nil, want %v", tt.want)
			} else if got != tt.want {
				t.Errorf("FindLegacy() = %v, want %v", got, tt.want)
			}
		})
		os.RemoveAll(tmpdir)
	}
}

func TestExtractName(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), literal_9536)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		prepare string
		args    args
		want    string
	}{
		{
			name: "Complete front-matter",
			prepare: `---
name: Bug Report
about: This is how you report bugs
---

**Template contents**
`,
			args: args{
				filePath: tmpfile.Name(),
			},
			want: "Bug Report",
		},
		{
			name: "Incomplete front-matter",
			prepare: `---
about: This is how you report bugs
---
`,
			args: args{
				filePath: tmpfile.Name(),
			},
			want: path.Base(tmpfile.Name()),
		},
		{
			name:    "No front-matter",
			prepare: `name: This is not yaml!`,
			args: args{
				filePath: tmpfile.Name(),
			},
			want: path.Base(tmpfile.Name()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.WriteFile(tmpfile.Name(), []byte(tt.prepare), 0600)
			if got := ExtractName(tt.args.filePath); got != tt.want {
				t.Errorf("ExtractName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractContents(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), literal_9536)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		prepare string
		args    args
		want    string
	}{
		{
			name: "Has front-matter",
			prepare: `---
name: Bug Report
---


Template contents
---
More of template
`,
			args: args{
				filePath: tmpfile.Name(),
			},
			want: `Template contents
---
More of template
`,
		},
		{
			name: "No front-matter",
			prepare: `Template contents
---
More of template
---
Even more
`,
			args: args{
				filePath: tmpfile.Name(),
			},
			want: `Template contents
---
More of template
---
Even more
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.WriteFile(tmpfile.Name(), []byte(tt.prepare), 0600)
			if got := ExtractContents(tt.args.filePath); string(got) != tt.want {
				t.Errorf("ExtractContents() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

const literal_9536 = "gh-cli"

const literal_4195 = "README.md"

const literal_7506 = "issue_template.md"

const literal_4830 = ".github/issue_template.md"

const literal_8065 = "docs/issue_template.md"

const literal_3056 = "ISSUE_TEMPLATE.md"

const literal_0634 = "docs/ISSUE_TEMPLATE/abc.md"

const literal_4710 = ".github/PULL_REQUEST_TEMPLATE.md"
