package format

import (
	"github.com/jialequ/mplb/pkg/cmd/project/shared/templet"
)

func ProjectState(project templet.Project) string {
	if project.Closed {
		return "closed"
	}
	return "open"
}

func ColorForProjectState(project templet.Project) string {
	if project.Closed {
		return "gray"
	}
	return "green"
}
