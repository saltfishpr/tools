package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/saltfishpr/tools/pkg/mod"
)

var version string

var (
	dryRun  bool
	verbose bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "modclean <path>",
		Short:   "Clean up Go module download cache",
		Args:    cobra.ExactArgs(1),
		Version: version,

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			if info, err := os.Stat(dir); err == nil && !info.IsDir() {
				return errors.Errorf("expected a directory, got %s", dir)
			}
			return run(dir)
		},
	}

	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "perform a dry run without making changes")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "enable verbose output")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %+v\n", err)
	}
}

func run(dir string) error {
	modFiles, err := listGoModFiles(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to list go.mod files in %s", dir)
	}

	keep := make(map[string]struct{})
	for _, modFile := range modFiles {
		modules, err := mod.ListAllModules(filepath.Dir(modFile))
		if err != nil {
			return err
		}
		for _, m := range modules {
			keep[m.Path+"@"+m.Version] = struct{}{}
		}
	}

	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		cmd := exec.Command("go", "env", "GOMODCACHE")
		out, err := cmd.Output()
		if err != nil {
			return errors.Wrap(err, "failed to get GOMODCACHE")
		}
		gomodcache = strings.TrimSpace(string(out))
	}
	cacheModFiles, err := listGoModFiles(gomodcache)
	if err != nil {
		return errors.Wrapf(err, "failed to list go.mod files in GOMODCACHE %s", gomodcache)
	}
	var total int
	for _, modFile := range cacheModFiles {
		modDir := filepath.Dir(modFile)
		modVer := strings.TrimPrefix(modDir, gomodcache+"/")
		if _, ok := keep[modVer]; ok {
			continue
		}
		total++
		if dryRun {
			log.Infof("Would remove: %s", modDir)
		} else {
			if verbose {
				log.Infof("Removing: %s", modDir)
			}
			if err := os.Remove(modDir); err != nil {
				return errors.Wrapf(err, "failed to remove %s", modDir)
			}
		}
	}
	log.Infof("Total modules to remove: %d", total)
	return nil
}

func listGoModFiles(dir string) ([]string, error) {
	var gomodFiles []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			gomodFiles = append(gomodFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to walk directory %s", dir)
	}
	return gomodFiles, nil
}
