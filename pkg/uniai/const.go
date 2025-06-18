package uniai

const (
	API_BASEURL  = ""
	ModelDefault = "uniai01:7b"

	Byte = 1

	KiloByte = Byte * 1000
	MegaByte = KiloByte * 1000

	defaultTemperature = 0.25
	defaultTopK        = 40
	defaultTopP        = 0.95
)

var (
	// DefaultOptions is the default model options used for inference.
	DefaultOptions = map[string]interface{}{
		"temperature": defaultTemperature,
		"top_k":       defaultTopK,
		"top_p":       defaultTopP,
	}
)
