package main

import (
	"bytes"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/imports"
)

var version string

var (
	write  bool
	module string
	mode   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "sortimports [-w] [-m string] [--mode string] <project-path>",
		Short:   "Sort Go imports",
		Long:    "Sort Go imports into standard library, third-party, and local imports groups.",
		Args:    cobra.ExactArgs(1),
		Version: version,

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := args[0]

			if module == "" {
				modPath := filepath.Join(projectDir, "go.mod")
				modData, err := os.ReadFile(modPath)
				if err != nil {
					return errors.WithStack(err)
				}
				modFile, err := modfile.Parse("go.mod", modData, nil)
				if err != nil {
					return errors.WithStack(err)
				}
				module = modFile.Module.Mod.Path
			}

			log.Infof("current module is %s", module)

			filteredFiles, err := getFilesByMode()
			if err != nil {
				return err
			}

			return filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				if !isGoFile(info) {
					return nil
				}
				if mode != "" && !lo.Contains(filteredFiles, path) {
					return nil
				}
				if err := processGoFile(path, nil, os.Stdout, module); err != nil {
					return err
				}
				return nil
			})
		},
	}

	rootCmd.Flags().BoolVarP(&write, "write", "w", false, "write result to (source) file instead of stdout")
	rootCmd.Flags().StringVarP(&module, "module", "m", "", "specify the project module path manually")
	rootCmd.Flags().StringVar(&mode, "mode", "", "specify file selection mode (diff: changed files, staged: staged files)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %+v\n", err)
	}
}

type importGroup struct {
	name    string
	prefix  string
	imports []*dst.ImportSpec
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func processGoFile(filePath string, in io.Reader, out io.Writer, modulePath string) error {
	log.Info("Processing file", "file", filePath)

	if in == nil {
		f, err := os.Open(filePath)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		in = f
	}

	groups := []importGroup{
		{name: "Standard Library", prefix: ""},
		{name: "Third-Party Imports", prefix: ""},
		{name: "Local Imports", prefix: modulePath},
	}

	src, err := io.ReadAll(in)
	if err != nil {
		return errors.WithStack(err)
	}

	// Process imports using golang.org/x/tools/imports
	processed, err := imports.Process(filePath, src, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	f, err := decorator.Parse(processed)
	if err != nil {
		return errors.WithStack(err)
	}

	// 分类 imports
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		switch {
		case strings.HasPrefix(importPath, modulePath):
			groups[2].imports = append(groups[2].imports, imp)
		case isStandardLibraryImport(importPath):
			groups[0].imports = append(groups[0].imports, imp)
		default:
			groups[1].imports = append(groups[1].imports, imp)
		}
	}

	// 排序
	for _, group := range groups {
		sort.Slice(group.imports, func(i, j int) bool {
			pathI := strings.Trim(group.imports[i].Path.Value, `"`)
			pathJ := strings.Trim(group.imports[j].Path.Value, `"`)
			return pathI < pathJ
		})
	}

	// 处理 group 间的 empty line
	for i, group := range groups {
		if len(group.imports) == 0 {
			continue
		}
		for j, imp := range group.imports {
			// 组内无空行
			imp.Decs.NodeDecs.Before = dst.NewLine
			imp.Decs.NodeDecs.After = dst.NewLine
			if j == len(group.imports)-1 {
				// 组间有空行
				imp.Decs.NodeDecs.Before = dst.NewLine
				imp.Decs.NodeDecs.After = dst.EmptyLine
				if i == len(groups)-1 {
					// 最后一个组的最后一个 import 后无空行
					imp.Decs.NodeDecs.Before = dst.NewLine
					imp.Decs.NodeDecs.After = dst.NewLine
				}
			}
		}
	}

	f.Imports = slices.Concat(groups[0].imports, groups[1].imports, groups[2].imports)

	// 修改 ast
	newImportDecl := &dst.GenDecl{
		Tok: token.IMPORT,
		Specs: lo.Map(f.Imports, func(imp *dst.ImportSpec, _ int) dst.Spec {
			return imp
		}),
	}
	var newDecls []dst.Decl
	if len(f.Imports) > 0 {
		newDecls = append(newDecls, newImportDecl)
	}
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok && genDecl.Tok == token.IMPORT {
			// 跳过 import 语句，相当于删除
			continue
		}
		newDecls = append(newDecls, decl)
	}
	f.Decls = newDecls

	var res bytes.Buffer
	if err := decorator.Fprint(&res, f); err != nil {
		return errors.WithStack(err)
	}

	if bytes.Equal(src, res.Bytes()) {
		return nil
	}

	if write {
		var perm os.FileMode
		if fi, err := os.Stat(filePath); err == nil {
			perm = fi.Mode() & os.ModePerm
		}
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		out = f
	}

	if _, err := out.Write(res.Bytes()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func isStandardLibraryImport(importPath string) bool {
	return !strings.Contains(importPath, ".")
}

// 根据 mode 获取待处理的文件列表
func getFilesByMode() ([]string, error) {
	if mode == "" {
		return nil, nil
	}
	switch mode {
	case "staged":
		return getStagedFiles()
	case "diff":
		return getDiffFiles()
	default:
		return nil, errors.New("unsupported mode")
	}
}

func getDiffFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	files := strings.Split(string(output), "\n")
	return lo.Filter(files, func(file string, _ int) bool {
		return file != ""
	}), nil
}

func getStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	files := strings.Split(string(output), "\n")
	return lo.Filter(files, func(file string, _ int) bool {
		return file != ""
	}), nil
}
