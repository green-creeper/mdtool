package converter

import "github.com/green-creeper/mdtool/pkg/models"

// Converter is the interface that all converters must implement
type Converter interface {
	// Convert performs the conversion from input to output
	Convert(req *models.ConvertRequest) *models.ConvertResponse

	// Name returns the name of this converter
	Name() string

	// SupportedFormats returns the source and target formats
	SupportedFormats() (source, target string)
}
