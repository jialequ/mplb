package view

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/shlex"
	"github.com/jialequ/mplb/internal/browser"
	"github.com/jialequ/mplb/internal/config"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/cmd/run/shared"
	workflowShared "github.com/jialequ/mplb/pkg/cmd/workflow/shared"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/jialequ/mplb/pkg/httpmock"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdView(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		tty      bool
		wants    ViewOptions
		wantsErr bool
	}{
		{
			name:     "blank nontty",
			wantsErr: true,
		},
		{
			name: "blank tty",
			tty:  true,
			wants: ViewOptions{
				Prompt: true,
			},
		},
		{
			name: "web tty",
			tty:  true,
			cli:  "--web",
			wants: ViewOptions{
				Prompt: true,
				Web:    true,
			},
		},
		{
			name: "web nontty",
			cli:  "1234 --web",
			wants: ViewOptions{
				Web:   true,
				RunID: "1234",
			},
		},
		{
			name:     "disallow web and log",
			tty:      true,
			cli:      "-w --log",
			wantsErr: true,
		},
		{
			name:     "disallow log and log-failed",
			tty:      true,
			cli:      "--log --log-failed",
			wantsErr: true,
		},
		{
			name: "exit status",
			cli:  "--exit-status 1234",
			wants: ViewOptions{
				RunID:      "1234",
				ExitStatus: true,
			},
		},
		{
			name: "verbosity",
			cli:  "-v",
			tty:  true,
			wants: ViewOptions{
				Verbose: true,
				Prompt:  true,
			},
		},
		{
			name: "with arg nontty",
			cli:  "1234",
			wants: ViewOptions{
				RunID: "1234",
			},
		},
		{
			name: "job id passed",
			cli:  "--job 1234",
			wants: ViewOptions{
				JobID: "1234",
			},
		},
		{
			name: "log passed",
			tty:  true,
			cli:  "--log",
			wants: ViewOptions{
				Prompt: true,
				Log:    true,
			},
		},
		{
			name: "tolerates both run and job id",
			cli:  "1234 --job 4567",
			wants: ViewOptions{
				JobID: "4567",
			},
		},
		{
			name: "run id with attempt",
			cli:  "1234 --attempt 2",
			wants: ViewOptions{
				RunID:   "1234",
				Attempt: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.SetStdinTTY(tt.tty)
			ios.SetStdoutTTY(tt.tty)

			f := &cmdutil.Factory{
				IOStreams: ios,
				Config: func() (config.Config, error) {
					return config.NewBlankConfig(), nil
				},
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *ViewOptions
			cmd := NewCmdView(f, func(opts *ViewOptions) error {
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
				return
			}

			assert.NoError(t, err)

			assert.Equal(t, tt.wants.RunID, gotOpts.RunID)
			assert.Equal(t, tt.wants.Prompt, gotOpts.Prompt)
			assert.Equal(t, tt.wants.ExitStatus, gotOpts.ExitStatus)
			assert.Equal(t, tt.wants.Verbose, gotOpts.Verbose)
			assert.Equal(t, tt.wants.Attempt, gotOpts.Attempt)
		})
	}
}

func TestViewRun(t *testing.T) {
	tests := []struct {
		name        string
		httpStubs   func(*httpmock.Registry)
		promptStubs func(*prompter.MockPrompter)
		opts        *ViewOptions
		tty         bool
		wantErr     bool
		wantOut     string
		browsedURL  string
		errMsg      string
	}{
		{
			name: "associate with PR",
			tty:  true,
			opts: &ViewOptions{
				RunID:  "3",
				Prompt: false,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(`{"data": {
		"repository": {
			"pullRequests": {
				"nodes": [
					{"number": 2898,
						"headRepository": {
						"owner": {
						"login": "OWNER"
						},
												"name": "REPO"}}
				]}}}}`))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
			},
			wantOut: "\n✓ trunk CI OWNER/REPO#2898 · 3\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\n\nFor more information about the job, try: gh run view --job=10\nView this run on GitHub: https://github.com/runs/3\n",
		},
		{
			name: "associate with PR with attempt",
			tty:  true,
			opts: &ViewOptions{
				RunID:   "3",
				Attempt: 3,
				Prompt:  false,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_2981),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(`{"data": {
		"repository": {
			"pullRequests": {
				"nodes": [
					{"number": 2898,
						"headRepository": {
						"owner": {
						"login": "OWNER"
						},
												"name": "REPO"}}
				]}}}}`))
				reg.Register(
					httpmock.REST("GET", literal_0269),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
			},
			wantOut: "\n✓ trunk CI OWNER/REPO#2898 · 3 (Attempt #3)\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\n\nFor more information about the job, try: gh run view --job=10\nView this run on GitHub: https://github.com/runs/3/attempts/3\n",
		},
		{
			name: "exit status, failed run",
			opts: &ViewOptions{
				RunID:      "1234",
				ExitStatus: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_5409),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_0921),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_2157),
					httpmock.JSONResponse(shared.FailedJobAnnotations))
			},
			wantOut: "\nX trunk CI · 1234\nTriggered via push about 59 minutes ago\n\nJOBS\nX sad job in 4m34s (ID 20)\n  ✓ barf the quux\n  X quux the barf\n\nANNOTATIONS\nX the job is sad\nsad job: blaze.py#420\n\n\nTo see what failed, try: gh run view 1234 --log-failed\nView this run on GitHub: https://github.com/runs/1234\n",
			wantErr: true,
		},
		{
			name: "with artifacts",
			opts: &ViewOptions{
				RunID: "3",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.JSONResponse(map[string][]shared.Artifact{
						"artifacts": {
							shared.Artifact{Name: "artifact-1", Expired: false},
							shared.Artifact{Name: "artifact-2", Expired: true},
							shared.Artifact{Name: "artifact-3", Expired: false},
						},
					}))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: heredoc.Doc(`

				✓ trunk CI · 3
				Triggered via push about 59 minutes ago

				JOBS


				ARTIFACTS
				artifact-1
				artifact-2 (expired)
				artifact-3

				For more information about a job, try: gh run view --job=<job-id>
				View this run on GitHub: https://github.com/runs/3
			`),
		},
		{
			name: "with artifacts and attempt",
			opts: &ViewOptions{
				RunID:   "3",
				Attempt: 3,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_2981),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.JSONResponse(map[string][]shared.Artifact{
						"artifacts": {
							shared.Artifact{Name: "artifact-1", Expired: false},
							shared.Artifact{Name: "artifact-2", Expired: true},
							shared.Artifact{Name: "artifact-3", Expired: false},
						},
					}))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_0269),
					httpmock.JSONResponse(shared.JobsPayload{}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: heredoc.Doc(`

				✓ trunk CI · 3 (Attempt #3)
				Triggered via push about 59 minutes ago

				JOBS


				ARTIFACTS
				artifact-1
				artifact-2 (expired)
				artifact-3

				For more information about a job, try: gh run view --job=<job-id>
				View this run on GitHub: https://github.com/runs/3/attempts/3
			`),
		},
		{
			name: "exit status, successful run",
			opts: &ViewOptions{
				RunID:      "3",
				ExitStatus: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: "\n✓ trunk CI · 3\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\n\nFor more information about the job, try: gh run view --job=10\nView this run on GitHub: https://github.com/runs/3\n",
		},
		{
			name: "exit status, successful run, with attempt",
			opts: &ViewOptions{
				RunID:      "3",
				Attempt:    3,
				ExitStatus: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_2981),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_0269),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: "\n✓ trunk CI · 3 (Attempt #3)\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\n\nFor more information about the job, try: gh run view --job=10\nView this run on GitHub: https://github.com/runs/3/attempts/3\n",
		},
		{
			name: "verbose",
			tty:  true,
			opts: &ViewOptions{
				RunID:   "1234",
				Prompt:  false,
				Verbose: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_5409),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_0921),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_2157),
					httpmock.JSONResponse(shared.FailedJobAnnotations))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: "\nX trunk CI · 1234\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\n  ✓ fob the barz\n  ✓ barz the fob\nX sad job in 4m34s (ID 20)\n  ✓ barf the quux\n  X quux the barf\n\nANNOTATIONS\nX the job is sad\nsad job: blaze.py#420\n\n\nTo see what failed, try: gh run view 1234 --log-failed\nView this run on GitHub: https://github.com/runs/1234\n",
		},
		{
			name: "prompts for choice, one job",
			tty:  true,
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_5179)
					})
			},
			opts: &ViewOptions{
				Prompt: true,
			},
			wantOut: "\n✓ trunk CI · 3\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\n\nFor more information about the job, try: gh run view --job=10\nView this run on GitHub: https://github.com/runs/3\n",
		},
		{
			name: "interactive with log",
			tty:  true,
			opts: &ViewOptions{
				Prompt: true,
				Log:    true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_8517),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_5179)
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_3762)
					})
			},
			wantOut: coolJobRunLogOutput,
		},
		{
			name: "interactive with log and attempt",
			tty:  true,
			opts: &ViewOptions{
				Prompt:  true,
				Attempt: 3,
				Log:     true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_2981),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_0269),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/3/attempts/3/logs"),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_5179)
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_3762)
					})
			},
			wantOut: coolJobRunLogOutput,
		},
		{
			name: "noninteractive with log",
			opts: &ViewOptions{
				JobID: "10",
				Log:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_9826),
					httpmock.JSONResponse(shared.SuccessfulJob))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_8517),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: coolJobRunLogOutput,
		},
		{
			name: "noninteractive with log and attempt",
			opts: &ViewOptions{
				JobID:   "10",
				Attempt: 3,
				Log:     true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_9826),
					httpmock.JSONResponse(shared.SuccessfulJob))
				reg.Register(
					httpmock.REST("GET", literal_2981),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/3/attempts/3/logs"),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: coolJobRunLogOutput,
		},
		{
			name: "interactive with run log",
			tty:  true,
			opts: &ViewOptions{
				Prompt: true,
				Log:    true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_8517),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_5179)
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_1629)
					})
			},
			wantOut: expectedRunLogOutput,
		},
		{
			name: "noninteractive with run log",
			tty:  true,
			opts: &ViewOptions{
				RunID: "3",
				Log:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_8517),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: expectedRunLogOutput,
		},
		{
			name: "interactive with log-failed",
			tty:  true,
			opts: &ViewOptions{
				Prompt:    true,
				LogFailed: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_0921),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9230),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return 4, nil
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_4692)
					})
			},
			wantOut: quuxTheBarfLogOutput,
		},
		{
			name: "interactive with log-failed with attempt",
			tty:  true,
			opts: &ViewOptions{
				Prompt:    true,
				Attempt:   3,
				LogFailed: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/1234/attempts/3"),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/1234/attempts/3/jobs"),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/1234/attempts/3/logs"),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return 4, nil
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_4692)
					})
			},
			wantOut: quuxTheBarfLogOutput,
		},
		{
			name: "noninteractive with log-failed",
			opts: &ViewOptions{
				JobID:     "20",
				LogFailed: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/jobs/20"),
					httpmock.JSONResponse(shared.FailedJob))
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_9230),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: quuxTheBarfLogOutput,
		},
		{
			name: "interactive with run log-failed",
			tty:  true,
			opts: &ViewOptions{
				Prompt:    true,
				LogFailed: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_0921),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9230),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return 4, nil
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_1629)
					})
			},
			wantOut: quuxTheBarfLogOutput,
		},
		{
			name: "noninteractive with run log-failed",
			tty:  true,
			opts: &ViewOptions{
				RunID:     "1234",
				LogFailed: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_0921),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9230),
					httpmock.FileResponse(literal_8624))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: quuxTheBarfLogOutput,
		},
		{
			name: "run log but run is not done",
			tty:  true,
			opts: &ViewOptions{
				RunID: "2",
				Log:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/2"),
					httpmock.JSONResponse(shared.TestRun(2, shared.InProgress, "")))
				reg.Register(
					httpmock.REST("GET", "runs/2/jobs"),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantErr: true,
			errMsg:  "run 2 is still in progress; logs will be available when it is complete",
		},
		{
			name: "job log but job is not done",
			tty:  true,
			opts: &ViewOptions{
				JobID: "20",
				Log:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/jobs/20"),
					httpmock.JSONResponse(shared.Job{
						ID:     20,
						Status: shared.InProgress,
						RunID:  2,
					}))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/2"),
					httpmock.JSONResponse(shared.TestRun(2, shared.InProgress, "")))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantErr: true,
			errMsg:  "job 20 is still in progress; logs will be available when it is complete",
		},
		{
			name: "noninteractive with job",
			opts: &ViewOptions{
				JobID: "10",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_9826),
					httpmock.JSONResponse(shared.SuccessfulJob))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: "\n✓ trunk CI · 3\nTriggered via push about 59 minutes ago\n\n✓ cool job in 4m34s (ID 10)\n  ✓ fob the barz\n  ✓ barz the fob\n\nTo see the full job log, try: gh run view --log --job=10\nView this run on GitHub: https://github.com/runs/3\n",
		},
		{
			name: "interactive, multiple jobs, choose all jobs",
			tty:  true,
			opts: &ViewOptions{
				Prompt: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5247),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_2157),
					httpmock.JSONResponse(shared.FailedJobAnnotations))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_5179)
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_1629)
					})
			},
			wantOut: "\n✓ trunk CI · 3\nTriggered via push about 59 minutes ago\n\nJOBS\n✓ cool job in 4m34s (ID 10)\nX sad job in 4m34s (ID 20)\n  ✓ barf the quux\n  X quux the barf\n\nANNOTATIONS\nX the job is sad\nsad job: blaze.py#420\n\n\nFor more information about a job, try: gh run view --job=<job-id>\nView this run on GitHub: https://github.com/runs/3\n",
		},
		{
			name: "interactive, multiple jobs, choose specific jobs",
			tty:  true,
			opts: &ViewOptions{
				Prompt: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6089),
					httpmock.JSONResponse(shared.RunsPayload{
						WorkflowRuns: shared.TestRuns,
					}))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.SuccessfulJob,
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_6293),
					httpmock.JSONResponse([]shared.Annotation{}))
				reg.Register(
					httpmock.REST("GET", literal_2683),
					httpmock.JSONResponse(workflowShared.WorkflowsPayload{
						Workflows: []workflowShared.Workflow{
							shared.TestWorkflow,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			promptStubs: func(pm *prompter.MockPrompter) {
				pm.RegisterSelect(literal_5602,
					[]string{literal_4581, literal_9345, literal_5179, literal_4581, literal_4581, literal_6059, literal_6059, literal_9345, literal_9345, literal_4581},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_5179)
					})
				pm.RegisterSelect(literal_3468,
					[]string{literal_1629, literal_3762, literal_4692},
					func(_, _ string, opts []string) (int, error) {
						return prompter.IndexFor(opts, literal_3762)
					})
			},
			wantOut: "\n✓ trunk CI · 3\nTriggered via push about 59 minutes ago\n\n✓ cool job in 4m34s (ID 10)\n  ✓ fob the barz\n  ✓ barz the fob\n\nTo see the full job log, try: gh run view --log --job=10\nView this run on GitHub: https://github.com/runs/3\n",
		},
		{
			name: "web run",
			tty:  true,
			opts: &ViewOptions{
				RunID: "3",
				Web:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			browsedURL: "https://github.com/runs/3",
			wantOut:    "Opening github.com/runs/3 in your browser.\n",
		},
		{
			name: "web job",
			tty:  true,
			opts: &ViewOptions{
				JobID: "10",
				Web:   true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_9826),
					httpmock.JSONResponse(shared.SuccessfulJob))
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			browsedURL: "https://github.com/jobs/10?check_suite_focus=true",
			wantOut:    "Opening github.com/jobs/10 in your browser.\n",
		},
		{
			name: "hide job header, failure",
			tty:  true,
			opts: &ViewOptions{
				RunID: "123",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123"),
					httpmock.JSONResponse(shared.TestRun(123, shared.Completed, shared.Failure)))
				reg.Register(
					httpmock.REST("GET", "runs/123/jobs"),
					httpmock.JSONResponse(shared.JobsPayload{Jobs: []shared.Job{}}))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123/artifacts"),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: "\nX trunk CI · 123\nTriggered via push about 59 minutes ago\n\nX This run likely failed because of a workflow file issue.\n\nFor more information, see: https://github.com/runs/123\n",
		},
		{
			name: "hide job header, startup_failure",
			tty:  true,
			opts: &ViewOptions{
				RunID: "123",
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123"),
					httpmock.JSONResponse(shared.TestRun(123, shared.Completed, shared.StartupFailure)))
				reg.Register(
					httpmock.REST("GET", "runs/123/jobs"),
					httpmock.JSONResponse(shared.JobsPayload{Jobs: []shared.Job{}}))
				reg.Register(
					httpmock.REST("GET", "repos/OWNER/REPO/actions/runs/123/artifacts"),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
			},
			wantOut: "\nX trunk CI · 123\nTriggered via push about 59 minutes ago\n\nX This run likely failed because of a workflow file issue.\n\nFor more information, see: https://github.com/runs/123\n",
		},
		{
			name: "Fetches all of a run's jobs with --json flag",
			opts: &ViewOptions{
				RunID: "3",
				Exporter: shared.MakeTestExporter(
					[]string{"jobs"},
					func(io *iostreams.IOStreams, data interface{}) error {
						run, ok := data.(*shared.Run)
						if !ok {
							return fmt.Errorf("expected data type *shared.Run")
						}
						fmt.Fprintf(io.Out, "fetched %d jobs\n", len(run.Jobs))
						return nil
					},
				),
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_3609),
					httpmock.JSONResponse(shared.SuccessfulRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.QueryMatcher("GET", literal_5693, url.Values{"per_page": []string{"100"}}),
					httpmock.WithHeader(
						httpmock.StringResponse(`{"jobs":[{},{},{}]}`),
						"Link",
						`<https://api.github.com/runs/3/jobs?page=2>; rel="next", <https://api.github.com/runs/3/jobs?page=2>; rel="last"`),
				)
				reg.Register(
					httpmock.REST("GET", literal_5693),
					httpmock.StringResponse(`{"jobs":[{},{}]}`))
			},
			wantOut: "fetched 5 jobs\n",
		},
		{
			name: "Returns error when failing to get annotations",
			opts: &ViewOptions{
				RunID:      "1234",
				ExitStatus: true,
			},
			httpStubs: func(reg *httpmock.Registry) {
				reg.Register(
					httpmock.REST("GET", literal_6741),
					httpmock.JSONResponse(shared.FailedRun))
				reg.Register(
					httpmock.REST("GET", literal_9748),
					httpmock.JSONResponse(shared.TestWorkflow))
				reg.Register(
					httpmock.REST("GET", literal_5409),
					httpmock.StringResponse(`{}`))
				reg.Register(
					httpmock.GraphQL(`query PullRequestForRun`),
					httpmock.StringResponse(``))
				reg.Register(
					httpmock.REST("GET", literal_0921),
					httpmock.JSONResponse(shared.JobsPayload{
						Jobs: []shared.Job{
							shared.FailedJob,
						},
					}))
				reg.Register(
					httpmock.REST("GET", literal_2157),
					httpmock.StatusStringResponse(500, "internal server error"),
				)
			},
			errMsg:  "failed to get annotations: HTTP 500 (https://api.github.com/repos/OWNER/REPO/check-runs/20/annotations)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		reg := &httpmock.Registry{}
		tt.httpStubs(reg)
		tt.opts.HttpClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}

		tt.opts.Now = func() time.Time {
			notnow, _ := time.Parse("2006-01-02 15:04:05", "2021-02-23 05:50:00")
			return notnow
		}

		ios, _, stdout, _ := iostreams.Test()
		ios.SetStdoutTTY(tt.tty)
		tt.opts.IO = ios
		tt.opts.BaseRepo = func() (ghrepo.Interface, error) {
			return ghrepo.FromFullName("OWNER/REPO")
		}

		pm := prompter.NewMockPrompter(t)
		tt.opts.Prompter = pm
		if tt.promptStubs != nil {
			tt.promptStubs(pm)
		}

		browser := &browser.Stub{}
		tt.opts.Browser = browser
		tt.opts.RunLogCache = RunLogCache{
			cacheDir: t.TempDir(),
		}

		t.Run(tt.name, func(t *testing.T) {
			err := runView(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
				if !tt.opts.ExitStatus {
					return
				}
			}
			if !tt.opts.ExitStatus {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantOut, stdout.String())
			if tt.browsedURL != "" {
				assert.Equal(t, tt.browsedURL, browser.BrowsedURL())
			}
			reg.Verify(t)
		})
	}
}

var barfTheFobLogOutput = heredoc.Doc(`
cool job	barz the fob	log line 1
cool job	barz the fob	log line 2
cool job	barz the fob	log line 3
`)

var fobTheBarzLogOutput = heredoc.Doc(`
cool job	fob the barz	log line 1
cool job	fob the barz	log line 2
cool job	fob the barz	log line 3
`)

var barfTheQuuxLogOutput = heredoc.Doc(`
sad job	barf the quux	log line 1
sad job	barf the quux	log line 2
sad job	barf the quux	log line 3
`)

var quuxTheBarfLogOutput = heredoc.Doc(`
sad job	quux the barf	log line 1
sad job	quux the barf	log line 2
sad job	quux the barf	log line 3
`)

var coolJobRunLogOutput = fmt.Sprintf("%s%s", fobTheBarzLogOutput, barfTheFobLogOutput)
var sadJobRunLogOutput = fmt.Sprintf("%s%s", barfTheQuuxLogOutput, quuxTheBarfLogOutput)
var expectedRunLogOutput = fmt.Sprintf("%s%s", coolJobRunLogOutput, sadJobRunLogOutput)

func TestRunLog(t *testing.T) {
	t.Run("when the cache dir doesn't exist, exists return false", func(t *testing.T) {
		cacheDir := t.TempDir() + "/non-existent-dir"
		rlc := RunLogCache{cacheDir: cacheDir}

		exists, err := rlc.Exists("unimportant-key")
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("when no cache entry has been created, exists returns false", func(t *testing.T) {
		cacheDir := t.TempDir()
		rlc := RunLogCache{cacheDir: cacheDir}

		exists, err := rlc.Exists("unimportant-key")
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("when a cache entry has been created, exists returns true", func(t *testing.T) {
		cacheDir := t.TempDir()
		rlc := RunLogCache{cacheDir: cacheDir}

		contents := strings.NewReader("unimportant-content")
		require.NoError(t, rlc.Create("key", contents))

		exists, err := rlc.Exists("key")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("when the cache dir doesn't exist, creating a cache entry creates it", func(t *testing.T) {
		cacheDir := t.TempDir() + "/non-existent-dir"
		rlc := RunLogCache{cacheDir: cacheDir}

		contents := strings.NewReader("unimportant-content")
		require.NoError(t, rlc.Create("key", contents))

		require.DirExists(t, cacheDir)
	})

	t.Run("when a cache entry has been created, reading it returns its contents", func(t *testing.T) {
		cacheDir := t.TempDir()
		rlc := RunLogCache{cacheDir: cacheDir}

		f, err := os.Open(literal_8624)
		require.NoError(t, err)
		defer f.Close()

		require.NoError(t, rlc.Create("key", f))

		zipReader, err := rlc.Open("key")
		require.NoError(t, err)
		defer zipReader.Close()
		require.NotEmpty(t, zipReader.File)
	})
}

const literal_3609 = "repos/OWNER/REPO/actions/runs/3"

const literal_9748 = "repos/OWNER/REPO/actions/workflows/123"

const literal_5247 = "repos/OWNER/REPO/actions/runs/3/artifacts"

const literal_5693 = "runs/3/jobs"

const literal_6293 = "repos/OWNER/REPO/check-runs/10/annotations"

const literal_2981 = "repos/OWNER/REPO/actions/runs/3/attempts/3"

const literal_0269 = "repos/OWNER/REPO/actions/runs/3/attempts/3/jobs"

const literal_6741 = "repos/OWNER/REPO/actions/runs/1234"

const literal_5409 = "repos/OWNER/REPO/actions/runs/1234/artifacts"

const literal_0921 = "runs/1234/jobs"

const literal_2157 = "repos/OWNER/REPO/check-runs/20/annotations"

const literal_6089 = "repos/OWNER/REPO/actions/runs"

const literal_2683 = "repos/OWNER/REPO/actions/workflows"

const literal_5602 = "Select a workflow run"

const literal_4581 = "X cool commit, CI (trunk) Feb 23, 2021"

const literal_9345 = "* cool commit, CI (trunk) Feb 23, 2021"

const literal_5179 = "✓ cool commit, CI (trunk) Feb 23, 2021"

const literal_6059 = "- cool commit, CI (trunk) Feb 23, 2021"

const literal_8517 = "repos/OWNER/REPO/actions/runs/3/logs"

const literal_8624 = "./fixtures/run_log.zip"

const literal_3468 = "View a specific job in this run?"

const literal_1629 = "View all jobs in this run"

const literal_3762 = "✓ cool job"

const literal_4692 = "X sad job"

const literal_9826 = "repos/OWNER/REPO/actions/jobs/10"

const literal_9230 = "repos/OWNER/REPO/actions/runs/1234/logs"

const literal_0368 = "cool job"

const literal_3657 = "fob the barz"
