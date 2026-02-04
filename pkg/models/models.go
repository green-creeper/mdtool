package models

import "io"

// ConvertRequest represents a conversion request
type ConvertRequest struct {
	Input   io.Reader
	Output  io.Writer
	Options map[string]interface{}
}

// ConvertResponse represents the result of a conversion
type ConvertResponse struct {
	Success  bool
	Error    error
	Metadata map[string]string
}
