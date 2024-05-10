package run

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdRun(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		tty      bool
		wants    RunOptions
		wantsErr bool
		errMsg   string
		stdin    string
	}{
		{
			name:     "blank nontty",
			wantsErr: true,
			errMsg:   "workflow ID, name, or filename required when not running interactively",
		},
		{
			name: "blank tty",
			tty:  true,
			wants: RunOptions{
				Prompt: true,
			},
		},
		{
			name: "ref flag",
			tty:  true,
			cli:  "--ref 12345abc",
			wants: RunOptions{
				Prompt: true,
				Ref:    "12345abc",
			},
		},
		{
			name:     "both STDIN and input fields",
			stdin:    "some json",
			cli:      "workflow.yml -fhey=there --json",
			errMsg:   "only one of STDIN or -f/-F can be passed",
			wantsErr: true,
		},
		{
			name: "-f args",
			tty:  true,
			cli:  `workflow.yml -fhey=there -fname="dana scully"`,
			wants: RunOptions{
				Selector:  literal_5967,
				RawFields: []string{literal_0718, "name=dana scully"},
			},
		},
		{
			name: "-F args",
			tty:  true,
			cli:  `workflow.yml -Fhey=there -Fname="dana scully" -Ffile=@cool.txt`,
			wants: RunOptions{
				Selector:    literal_5967,
				MagicFields: []string{literal_0718, "name=dana scully", "file=@cool.txt"},
			},
		},
		{
			name: "-F/-f arg mix",
			tty:  true,
			cli:  `workflow.yml -fhey=there -Fname="dana scully" -Ffile=@cool.txt`,
			wants: RunOptions{
				Selector:    literal_5967,
				RawFields:   []string{literal_0718},
				MagicFields: []string{`name=dana scully`, "file=@cool.txt"},
			},
		},
		{
			name:  "json on STDIN",
			cli:   "workflow.yml --json",
			stdin: `{"cool":"yeah"}`,
			wants: RunOptions{
				JSON:      true,
				JSONInput: `{"cool":"yeah"}`,
				Selector:  literal_5967,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, stdin, _, _ := iostreams.Test()
			if tt.stdin == "" {
				ios.SetStdinTTY(tt.tty)
			} else {
				stdin.WriteString(tt.stdin)
			}
			ios.SetStdoutTTY(tt.tty)

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *RunOptions
			cmd := NewCmdRun(f, func(opts *RunOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
				return
			}

			assert.NoError(t, err)

			assert.Equal(t, tt.wants.Selector, gotOpts.Selector)
			assert.Equal(t, tt.wants.Prompt, gotOpts.Prompt)
			assert.Equal(t, tt.wants.JSONInput, gotOpts.JSONInput)
			assert.Equal(t, tt.wants.JSON, gotOpts.JSON)
			assert.Equal(t, tt.wants.Ref, gotOpts.Ref)
			assert.ElementsMatch(t, tt.wants.RawFields, gotOpts.RawFields)
			assert.ElementsMatch(t, tt.wants.MagicFields, gotOpts.MagicFields)
		})
	}
}

func TestMagicFieldValue(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "gh-test")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fmt.Fprint(f, "file contents")

	ios, _, _, _ := iostreams.Test()

	type args struct {
		v    string
		opts RunOptions
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "string",
			args:    args{v: "hello"},
			want:    "hello",
			wantErr: false,
		},
		{
			name: "file",
			args: args{
				v:    "@" + f.Name(),
				opts: RunOptions{IO: ios},
			},
			want:    "file contents",
			wantErr: false,
		},
		{
			name: "file error",
			args: args{
				v:    "@",
				opts: RunOptions{IO: ios},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := magicFieldValue(tt.args.v, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("magicFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindInputs(t *testing.T) {
	tests := []struct {
		name    string
		YAML    []byte
		wantErr bool
		errMsg  string
		wantOut []WorkflowInput
	}{
		{
			name:    "blank",
			YAML:    []byte{},
			wantErr: true,
			errMsg:  "invalid YAML file",
		},
		{
			name:    "no event specified",
			YAML:    []byte("name: workflow"),
			wantErr: true,
			errMsg:  "invalid workflow: no 'on' key",
		},
		{
			name:    "not workflow_dispatch",
			YAML:    []byte("name: workflow\non: pull_request"),
			wantErr: true,
			errMsg:  "unable to manually run a workflow without a workflow_dispatch event",
		},
		{
			name:    "bad inputs",
			YAML:    []byte("name: workflow\non:\n workflow_dispatch:\n  inputs: lol  "),
			wantErr: true,
			errMsg:  "could not decode workflow inputs: yaml: unmarshal errors:\n  line 4: cannot unmarshal !!str `lol` into map[string]run.WorkflowInput",
		},
		{
			name:    "short syntax",
			YAML:    []byte("name: workflow\non: workflow_dispatch"),
			wantOut: []WorkflowInput{},
		},
		{
			name:    "array of events",
			YAML:    []byte("name: workflow\non: [pull_request, workflow_dispatch]\n"),
			wantOut: []WorkflowInput{},
		},
		{
			name: "inputs",
			YAML: []byte(`name: workflow
on:
  workflow_dispatch:
    inputs:
      foo:
        required: true
        description: good foo
      bar:
        default: boo
      baz:
        description: it's baz
      quux:
        required: true
        default: "cool"
jobs:
  yell:
    runs-on: ubuntu-latest
    steps:
      - name: echo
        run: |
          echo "echo"`),
			wantOut: []WorkflowInput{
				{
					Name:    "bar",
					Default: "boo",
				},
				{
					Name:        "baz",
					Description: "it's baz",
				},
				{
					Name:        "foo",
					Required:    true,
					Description: "good foo",
				},
				{
					Name:     "quux",
					Required: true,
					Default:  "cool",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findInputs(tt.YAML)
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, tt.errMsg, err.Error())
				}
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantOut, result)
		})
	}

}

const literal_5967 = "workflow.yml"

const literal_0718 = "hey=there"
