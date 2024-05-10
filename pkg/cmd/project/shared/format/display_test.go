package format

import (
	"testing"

	"github.com/jialequ/mplb/pkg/cmd/project/shared/templet"
	"github.com/stretchr/testify/assert"
)

func TestProjectState(t *testing.T) {
	assert.Equal(t, "open", ProjectState(templet.Project{}))
	assert.Equal(t, "closed", ProjectState(templet.Project{Closed: true}))
}

func TestColorForProjectState(t *testing.T) {
	assert.Equal(t, "green", ColorForProjectState(templet.Project{}))
	assert.Equal(t, "gray", ColorForProjectState(templet.Project{Closed: true}))
}
