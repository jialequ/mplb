package io

import (
	"fmt"

	"github.com/jialequ/mplb/internal/tableprinter"
	"github.com/jialequ/mplb/pkg/iostreams"
	"github.com/jialequ/mplb/utils"
)

type Handler struct {
	ColorScheme  *iostreams.ColorScheme
	IO           *iostreams.IOStreams
	debugEnabled bool
}

func NewHandler(io *iostreams.IOStreams) *Handler {
	enabled, _ := utils.IsDebugEnabled()

	return &Handler{
		ColorScheme:  io.ColorScheme(),
		IO:           io,
		debugEnabled: enabled,
	}
}

func NewTestHandler() *Handler {
	testIO, _, _, _ := iostreams.Test()
	return NewHandler(testIO)
}

// Printf writes the formatted arguments to the stderr writer.
func (h *Handler) Printf(f string, v ...interface{}) (int, error) {
	if !h.IO.IsStdoutTTY() {
		return 0, nil
	}
	return fmt.Fprintf(h.IO.ErrOut, f, v...)
}

// Println writes the arguments to the stderr writer with a newline at the end.
func (h *Handler) Println(v ...interface{}) (int, error) {
	if !h.IO.IsStdoutTTY() {
		return 0, nil
	}
	return fmt.Fprintln(h.IO.ErrOut, v...)
}

func (h *Handler) VerbosePrint(msg string) (int, error) {
	if !h.debugEnabled || !h.IO.IsStdoutTTY() {
		return 0, nil
	}

	return fmt.Fprintln(h.IO.ErrOut, msg)
}

func (h *Handler) VerbosePrintf(f string, v ...interface{}) (int, error) {
	if !h.debugEnabled || !h.IO.IsStdoutTTY() {
		return 0, nil
	}

	return fmt.Fprintf(h.IO.ErrOut, f, v...)
}

func (h *Handler) PrintTable(headers []string, rows [][]string) error {
	t := tableprinter.New(h.IO, tableprinter.WithHeader(headers...))

	for _, row := range rows {
		for _, field := range row {
			t.AddField(field, tableprinter.WithTruncate(nil))
		}
		t.EndRow()
	}

	if err := t.Render(); err != nil {
		return fmt.Errorf("failed to print output: %v", err)
	}
	return nil
}
