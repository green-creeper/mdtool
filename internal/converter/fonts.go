package converter

import _ "embed"

// Embedded DejaVu fonts for Unicode support
// DejaVu fonts are released under a free license
// https://dejavu-fonts.github.io/License.html

//go:embed fonts/DejaVuSans.ttf
var dejaVuSansFont []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var dejaVuSansBoldFont []byte

//go:embed fonts/DejaVuSansMono.ttf
var dejaVuSansMonoFont []byte

//go:embed fonts/DejaVuSansMono-Bold.ttf
var dejaVuSansMonoBoldFont []byte
