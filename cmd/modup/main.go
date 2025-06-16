package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"

	"github.com/saltfishpr/tools/pkg/mod"
	"github.com/saltfishpr/tools/pkg/util"
)

var version string

var (
	write     bool
	indirect  bool
	proxy     string
	goVersion string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "modup [-w] [--indirect] [--proxy string] [-go string] <path>",
		Short:   "Manage go module dependencies",
		Long:    "Upgrade or Downgrade dependencies to the latest version compatible with the target Go version.",
		Args:    cobra.ExactArgs(1),
		Version: version,

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			if info, err := os.Stat(target); err == nil && info.IsDir() {
				target = filepath.Join(target, "go.mod")
			}
			if proxy != "" {
				mod.SetGoProxy(proxy)
			}
			return run(target, nil, os.Stdout)
		},
	}

	rootCmd.Flags().BoolVarP(&write, "write", "w", false, "write result to (source) file instead of stdout")
	rootCmd.Flags().BoolVar(&indirect, "indirect", false, "upgrade indirect dependencies")
	rootCmd.Flags().StringVar(&proxy, "proxy", "", "use the specified proxy instead of reading from the environment")
	rootCmd.Flags().StringVar(&goVersion, "go", "", "target Go version to use for compatibility checks")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %+v\n", err)
	}
}

func run(filename string, in io.Reader, out io.Writer) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		in = f
	}

	data, err := io.ReadAll(in)
	if err != nil {
		return errors.WithStack(err)
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	currentGoVersion := goVersion
	if currentGoVersion == "" {
		currentGoVersion = f.Go.Version
	}
	f.Go.Syntax.Token[1] = currentGoVersion

	_, cMinor, _, err := util.ParseGoVersion(currentGoVersion)
	if err != nil {
		return errors.WithStack(err)
	}
	if cMinor < 21 {
		if f.Toolchain != nil {
			f.Toolchain.Syntax.Token = nil // remove toolchain
		}
	}

	var deps []*modfile.Require
	for _, req := range f.Require {
		if req.Indirect {
			if indirect {
				deps = append(deps, req)
			}
			continue
		}
		deps = append(deps, req)
	}

	for _, dep := range deps {
		latest, err := findLatestCompatible(dep.Mod.Path, dep.Mod.Version, currentGoVersion)
		if err != nil {
			continue
		}
		if latest != "" && latest != dep.Mod.Version {
			log.Infof("updating %s from %s to %s", dep.Mod.Path, dep.Mod.Version, latest)
			dep.Syntax.Token[1] = latest
		}
	}

	content, err := f.Format()
	if err != nil {
		return errors.WithStack(err)
	}

	if bytes.Equal(data, content) {
		log.Info("no changes")
		return nil
	}

	if write {
		var perm os.FileMode
		if fi, err := os.Stat(filename); err == nil {
			perm = fi.Mode() & os.ModePerm
		}
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		out = f
	}

	if _, err := out.Write(content); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func findLatestCompatible(depModPath, depModVersion, goVersion string) (string, error) {
	versions, err := mod.ListVersions(depModPath)
	if err != nil {
		return "", err
	}
	depModMajor := semver.Major(depModVersion)
	if depModMajor == "" {
		return "", errors.New("invalid version")
	}

	for _, ver := range versions {
		if !semver.IsValid(ver) {
			log.Warnf("invalid version %s for %s", ver, depModPath)
			continue
		}

		if semver.Major(ver) != depModMajor {
			break // 大版本不匹配
		}

		if semver.Prerelease(ver) != "" {
			continue // 非稳定版本
		}

		f, err := mod.GetModFile(depModPath, ver)
		if err != nil {
			log.Errorf("get mod file %s@%s failed: %v", depModPath, ver, err)
			continue
		}

		if f.Go == nil {
			log.Warnf("mod %s@%s has no go version, skipping this package", depModPath, ver)
			return depModVersion, nil
		}

		if isCompatible(goVersion, f.Go.Version) {
			return ver, nil
		}
	}
	return "", nil
}

func isCompatible(currentGoVer, modGoVer string) bool {
	cMajor, cMinor, _, err1 := util.ParseGoVersion(currentGoVer)
	mMajor, mMinor, _, err2 := util.ParseGoVersion(modGoVer)
	if err1 != nil || err2 != nil {
		return false // 解析失败认为不兼容
	}
	return cMajor == mMajor && cMinor >= mMinor
}
