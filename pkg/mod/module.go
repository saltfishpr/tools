package mod

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"time"
)

type Module struct {
	Path       string       // module path
	Query      string       // version query corresponding to this version
	Version    string       // module version
	Versions   []string     // available module versions
	Replace    *Module      // replaced by this module
	Time       *time.Time   // time version was created
	Update     *Module      // available update (with -u)
	Main       bool         // is this the main module?
	Indirect   bool         // module is only indirectly needed by main module
	Dir        string       // directory holding local copy of files, if any
	GoMod      string       // path to go.mod file describing module, if any
	GoVersion  string       // go version used in module
	Retracted  []string     // retraction information, if any (with -retracted or -u)
	Deprecated string       // deprecation message, if any (with -u)
	Error      *ModuleError // error loading module
	Sum        string       // checksum for path, version (as in go.sum)
	GoModSum   string       // checksum for go.mod (as in go.sum)
	Origin     any          // provenance of module
	Reuse      bool         // reuse of old module info is safe
}

type ModuleError struct {
	Err string // the error itself
}

func ListAllModules(dir string) ([]Module, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(out))
	var modules []Module
	for decoder.More() {
		var mod Module
		if err := decoder.Decode(&mod); err != nil {
			return nil, err
		}
		if mod.Main {
			continue
		}
		modules = append(modules, mod)
	}
	return modules, nil
}
