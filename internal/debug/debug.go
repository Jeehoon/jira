package debug

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	Enabled = false
	enc     = json.NewEncoder(os.Stderr)
)

func init() {
	//enc.SetIndent("", "  ")
}

func Printf(f string, args ...any) {
	if !Enabled {
		return
	}
	fmt.Fprintf(os.Stderr, f+"\n\n", args...)
}

func DumpJson(v any) {
	if !Enabled {
		return
	}
	enc.Encode(v)
	fmt.Fprintf(os.Stderr, "\n")
}
