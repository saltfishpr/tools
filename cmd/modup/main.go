package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"

	"github/saltfishpr/tools/pkg/mod"
	"github/saltfishpr/tools/pkg/util"
)

var (
	write    bool
	verbose  bool
	indirect bool
)

func init() {
	flag.BoolVar(&write, "w", false, "write result to (source) file instead of stdout")
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.BoolVar(&indirect, "indirect", false, "include indirect dependencies")
	flag.CommandLine.Init("modup", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: modup [flags] <mod-file>\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("usage: modup [flags] <mod-file>")
	}
	target := flag.Arg(0)
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		target = filepath.Join(target, "go.mod")
	}
	if err := run(target, nil, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(filename string, in io.Reader, out io.Writer) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	data, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return err
	}
	currentGoVersion := f.Go.Version

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
		log.Infof("checking %s %s", dep.Mod.Path, dep.Mod.Version)
		latest, err := findLatestCompatible(dep.Mod.Path, dep.Mod.Version, currentGoVersion)
		if err != nil {
			continue
		}
		if latest != "" && latest != dep.Mod.Version {
			dep.Syntax.Token[1] = latest
		}
	}

	content, err := f.Format()
	if err != nil {
		return err
	}

	if bytes.Equal(data, content) {
		if verbose {
			log.Info("no changes")
		}
		return nil
	}

	if write {
		var perm os.FileMode
		if fi, err := os.Stat(filename); err == nil {
			perm = fi.Mode() & os.ModePerm
		}
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}

	if _, err := out.Write(content); err != nil {
		return err
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
			continue
		}

		if semver.Major(ver) != depModMajor {
			continue
		}

		f, err := mod.GetModFile(depModPath, ver)
		if err != nil {
			log.Errorf("get mod file %s@%s failed: %v", depModPath, ver, err)
			continue
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
