package file

import (
	"github.com/uvasoftware/scanii-cli/internal/testutil"
)

var ts *testutil.Server

func init() {
	ts = testutil.StartServer()
}
