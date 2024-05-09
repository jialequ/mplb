package shared

import (
	"testing"

	"github.com/jialequ/mplb/api"
	"github.com/jialequ/mplb/internal/ghrepo"
	"github.com/jialequ/mplb/internal/prompter"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

type metadataFetcher struct {
	metadataResult *api.RepoMetadataResult
}

func (mf *metadataFetcher) RepoMetadataFetch(input api.RepoMetadataInput) (*api.RepoMetadataResult, error) {
	return mf.metadataResult, nil
}

func TestMetadataSurvey_selectAll(t *testing.T) {
	ios, _, stdout, stderr := iostreams.Test()

	repo := ghrepo.New("OWNER", "REPO")

	fetcher := &metadataFetcher{
		metadataResult: &api.RepoMetadataResult{
			AssignableUsers: []api.RepoAssignee{
				{Login: "hubot"},
				{Login: "monalisa"},
			},
			Labels: []api.RepoLabel{
				{Name: literal_1348},
				{Name: literal_2987},
			},
			Projects: []api.RepoProject{
				{Name: literal_3861},
				{Name: literal_3068},
			},
			Milestones: []api.RepoMilestone{
				{Title: "1.2 patch release"},
			},
		},
	}

	pm := prompter.NewMockPrompter(t)
	pm.RegisterMultiSelect("What would you like to add?",
		[]string{}, []string{"Reviewers", "Assignees", "Labels", "Projects", "Milestone"}, func(_ string, _, _ []string) ([]int, error) {
			return []int{0, 1, 2, 3, 4}, nil
		})
	pm.RegisterMultiSelect("Reviewers", []string{}, []string{"hubot", "monalisa"}, func(_ string, _, _ []string) ([]int, error) {
		return []int{1}, nil
	})
	pm.RegisterMultiSelect("Assignees", []string{}, []string{"hubot", "monalisa"}, func(_ string, _, _ []string) ([]int, error) {
		return []int{0}, nil
	})
	pm.RegisterMultiSelect("Labels", []string{}, []string{literal_1348, literal_2987}, func(_ string, _, _ []string) ([]int, error) {
		return []int{1}, nil
	})
	pm.RegisterMultiSelect("Projects", []string{}, []string{literal_3861, literal_3068}, func(_ string, _, _ []string) ([]int, error) {
		return []int{1}, nil
	})
	pm.RegisterSelect("Milestone", []string{"(none)", "1.2 patch release"}, func(_, _ string, _ []string) (int, error) {
		return 0, nil
	})

	state := &IssueMetadataState{
		Assignees: []string{"hubot"},
		Type:      PRMetadata,
	}
	err := MetadataSurvey(pm, ios, repo, fetcher, state)
	assert.NoError(t, err)

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())

	assert.Equal(t, []string{"hubot"}, state.Assignees)
	assert.Equal(t, []string{"monalisa"}, state.Reviewers)
	assert.Equal(t, []string{literal_2987}, state.Labels)
	assert.Equal(t, []string{literal_3068}, state.Projects)
	assert.Equal(t, []string{}, state.Milestones)
}

func TestMetadataSurvey_keepExisting(t *testing.T) {
	ios, _, stdout, stderr := iostreams.Test()

	repo := ghrepo.New("OWNER", "REPO")

	fetcher := &metadataFetcher{
		metadataResult: &api.RepoMetadataResult{
			Labels: []api.RepoLabel{
				{Name: literal_1348},
				{Name: literal_2987},
			},
			Projects: []api.RepoProject{
				{Name: literal_3861},
				{Name: literal_3068},
			},
		},
	}

	pm := prompter.NewMockPrompter(t)
	pm.RegisterMultiSelect("What would you like to add?", []string{}, []string{"Assignees", "Labels", "Projects", "Milestone"}, func(_ string, _, _ []string) ([]int, error) {
		return []int{1, 2}, nil
	})
	pm.RegisterMultiSelect("Labels", []string{}, []string{literal_1348, literal_2987}, func(_ string, _, _ []string) ([]int, error) {
		return []int{1}, nil
	})
	pm.RegisterMultiSelect("Projects", []string{}, []string{literal_3861, literal_3068}, func(_ string, _, _ []string) ([]int, error) {
		return []int{1}, nil
	})

	state := &IssueMetadataState{
		Assignees: []string{"hubot"},
	}
	err := MetadataSurvey(pm, ios, repo, fetcher, state)
	assert.NoError(t, err)

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "", stderr.String())

	assert.Equal(t, []string{"hubot"}, state.Assignees)
	assert.Equal(t, []string{literal_2987}, state.Labels)
	assert.Equal(t, []string{literal_3068}, state.Projects)
}

const literal_1348 = "help wanted"

const literal_2987 = "good first issue"

const literal_3861 = "Huge Refactoring"

const literal_3068 = "The road to 1.0"
