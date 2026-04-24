package assets

import (
	"embed"
	"encoding/hex"
	"io/fs"
	"strings"
)

//go:embed "samples"
var EmbeddedFiles embed.FS

func DecodedEICAR() string {
	encoded, err := fs.ReadFile(EmbeddedFiles, "samples/eicar.txt.b64")
	if err != nil {
		panic("read eicar fixture: " + err.Error())
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil {
		panic("decode eicar fixture: " + err.Error())
	}
	return string(decoded)
}
