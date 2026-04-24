package file

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/uvasoftware/scanii-cli/assets"
	"github.com/uvasoftware/scanii-cli/internal/testutil"
)

var ts *testutil.Server

var fakeMalwareSample = copyToTemp()

func copyToTemp() string {

	dir, err := os.MkdirTemp("", "fixture-")
	if err != nil {
		panic("mktemp: " + err.Error())
	}
	path := filepath.Join(dir, "sample")
	contents, err := fs.ReadFile(assets.EmbeddedFiles, "samples/malware")
	if err != nil {
		panic("read malware sample fixture: " + err.Error())
	}
	if err := os.WriteFile(path, contents, 0600); err != nil {
		panic("write malware sample fixture: " + err.Error())
	}
	return path
}

func init() {
	ts = testutil.StartServer()
}
