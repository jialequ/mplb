package list

import (
	"bytes"
	"testing"

	"github.com/google/shlex"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wants    ListOptions
		wantsErr string
	}{
		{
			name:  "no arguments",
			input: "",
			wants: ListOptions{
				Limit: 30,
				Order: "desc",
				Sort:  "last_accessed_at",
				Key:   "",
				Ref:   "",
			},
		},
		{
			name:  "with limit",
			input: "--limit 100",
			wants: ListOptions{
				Limit: 100,
				Order: "desc",
				Sort:  "last_accessed_at",
				Key:   "",
				Ref:   "",
			},
		},
		{
			name:     "invalid limit",
			input:    "-L 0",
			wantsErr: "invalid limit: 0",
		},
		{
			name:  "with sort",
			input: "--sort created_at",
			wants: ListOptions{
				Limit: 30,
				Order: "desc",
				Sort:  "created_at",
				Key:   "",
				Ref:   "",
			},
		},
		{
			name:  "with order",
			input: "--order asc",
			wants: ListOptions{
				Limit: 30,
				Order: "asc",
				Sort:  "last_accessed_at",
				Key:   "",
				Ref:   "",
			},
		},
		{
			name:  "with key",
			input: "--key cache-key-prefix-",
			wants: ListOptions{
				Limit: 30,
				Order: "desc",
				Sort:  "last_accessed_at",
				Key:   "cache-key-prefix-",
				Ref:   "",
			},
		},
		{
			name:  "with ref",
			input: "--ref refs/heads/main",
			wants: ListOptions{
				Limit: 30,
				Order: "desc",
				Sort:  "last_accessed_at",
				Key:   "",
				Ref:   literal_6218,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}
			argv, err := shlex.Split(tt.input)
			assert.NoError(t, err)
			var gotOpts *ListOptions
			cmd := NewCmdList(f, func(opts *ListOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantsErr != "" {
				assert.EqualError(t, err, tt.wantsErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.Limit, gotOpts.Limit)
			assert.Equal(t, tt.wants.Sort, gotOpts.Sort)
			assert.Equal(t, tt.wants.Order, gotOpts.Order)
			assert.Equal(t, tt.wants.Key, gotOpts.Key)
		})
	}
}

func TestHumanFileSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{
			name: "min bytes",
			size: 1,
			want: "1 B",
		},
		{
			name: "max bytes",
			size: 1023,
			want: "1023 B",
		},
		{
			name: "min kibibytes",
			size: 1024,
			want: "1.00 KiB",
		},
		{
			name: "max kibibytes",
			size: 1024*1024 - 1,
			want: "1023.99 KiB",
		},
		{
			name: "min mibibytes",
			size: 1024 * 1024,
			want: "1.00 MiB",
		},
		{
			name: "fractional mibibytes",
			size: 1024*1024*12 + 1024*350,
			want: "12.34 MiB",
		},
		{
			name: "max mibibytes",
			size: 1024*1024*1024 - 1,
			want: "1023.99 MiB",
		},
		{
			name: "min gibibytes",
			size: 1024 * 1024 * 1024,
			want: "1.00 GiB",
		},
		{
			name: "fractional gibibytes",
			size: 1024 * 1024 * 1024 * 1.5,
			want: "1.50 GiB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanFileSize(tt.size); got != tt.want {
				t.Errorf("humanFileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

const literal_6218 = "refs/heads/main"
