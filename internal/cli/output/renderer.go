package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Renderer struct {
	out    io.Writer
	format string
}

func New(out io.Writer, format string) *Renderer {
	return &Renderer{out: out, format: format}
}

func (r *Renderer) Format() string {
	return r.format
}

func (r *Renderer) WriteValue(value any, text string) error {
	switch r.format {
	case "text":
		_, err := fmt.Fprintln(r.out, text)
		return err
	case "json":
		encoder := json.NewEncoder(r.out)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "jsonl":
		return r.WriteJSONLine(value)
	default:
		return fmt.Errorf("unsupported output format %q", r.format)
	}
}

func (r *Renderer) WriteJSONLine(value any) error {
	encoder := json.NewEncoder(r.out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

type Log struct {
	ApplicationID string `json:"application_id,omitempty"`
	BuildID       string `json:"build_id,omitempty"`
	Time          string `json:"time,omitempty"`
	Text          string `json:"log"`
}

func (r *Renderer) WriteLog(log Log, streaming bool) error {
	if r.format == "text" {
		_, err := io.WriteString(r.out, log.Text)
		if err == nil && log.Text != "" && !strings.HasSuffix(log.Text, "\n") {
			_, err = io.WriteString(r.out, "\n")
		}
		return err
	}
	if streaming || r.format == "jsonl" {
		return r.WriteJSONLine(log)
	}
	return nil
}

func (r *Renderer) WriteString(value string) error {
	_, err := io.WriteString(r.out, value)
	return err
}

func (r *Renderer) Writef(format string, args ...any) error {
	_, err := fmt.Fprintf(r.out, format, args...)
	return err
}
