package core

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Writer struct {
	out    io.Writer
	format string
}

func New(out io.Writer, format string) *Writer {
	return &Writer{out: out, format: format}
}

func (w *Writer) Format() string {
	return w.format
}

func (w *Writer) WriteValue(value any, text string) error {
	switch w.format {
	case "text":
		_, err := fmt.Fprintln(w.out, text)
		return err
	case "json":
		encoder := json.NewEncoder(w.out)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "jsonl":
		return w.WriteJSONLine(value)
	default:
		return fmt.Errorf("unsupported output format %q", w.format)
	}
}

func (w *Writer) WriteJSONLine(value any) error {
	encoder := json.NewEncoder(w.out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

type Log struct {
	ApplicationID string `json:"application_id,omitempty"`
	BuildID       string `json:"build_id,omitempty"`
	Time          string `json:"time,omitempty"`
	Text          string `json:"log"`
}

func (w *Writer) WriteLog(log Log, streaming bool) error {
	if w.format == "text" {
		_, err := io.WriteString(w.out, log.Text)
		if err == nil && log.Text != "" && !strings.HasSuffix(log.Text, "\n") {
			_, err = io.WriteString(w.out, "\n")
		}
		return err
	}
	if streaming || w.format == "jsonl" {
		return w.WriteJSONLine(log)
	}
	return nil
}

func (w *Writer) WriteString(value string) error {
	_, err := io.WriteString(w.out, value)
	return err
}

func (w *Writer) Writef(format string, args ...any) error {
	_, err := fmt.Fprintf(w.out, format, args...)
	return err
}
