package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: which <program>")
		os.Exit(1)
	}

	name := os.Args[1]
	path := findExecutable(name)

	if path == "" {
		fmt.Fprintf(os.Stderr, "%s not found in PATH\n", name)
		os.Exit(1)
	}

	fmt.Println(path)
}

func getExtensions() []string {
	if runtime.GOOS != "windows" {
		return nil
	}

	pathExt := os.Getenv("PATHEXT")
	if pathExt == "" {
		return []string{".COM", ".EXE", ".BAT", ".CMD"}
	}

	exts := strings.Split(pathExt, ";")
	var result []string
	for _, ext := range exts {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, ext)
		}
	}
	return result
}

func isPath(name string) bool {
	return strings.ContainsAny(name, `/\`)
}

func findExecutable(name string) string {
	if isPath(name) {
		return findInDir(filepath.Dir(name), filepath.Base(name))
	}

	pathEnv := os.Getenv("PATH")

	var dirs []string

	if runtime.GOOS == "windows" {
		cwd, err := os.Getwd()
		if err == nil {
			dirs = append(dirs, cwd)
		}
	}

	if pathEnv != "" {
		dirs = append(dirs, filepath.SplitList(pathEnv)...)
	}

	for _, dir := range dirs {
		path := findInDir(dir, name)
		if path != "" {
			return path
		}
	}

	return ""
}

func findInDir(dir, name string) string {
	extensions := getExtensions()

	if len(extensions) > 0 {
		ext := strings.ToUpper(filepath.Ext(name))

		for _, e := range extensions {
			if ext == strings.ToUpper(e) {
				path := filepath.Join(dir, name)
				if isExecutable(path) {
					return normalizePath(path)
				}
				return ""
			}
		}

		for _, ext := range extensions {
			path := filepath.Join(dir, name+ext)
			if isExecutable(path) {
				return normalizePath(path)
			}
		}
	} else {
		path := filepath.Join(dir, name)
		if isExecutable(path) {
			return normalizePath(path)
		}
	}

	return ""
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	return true
}

func normalizePath(path string) string {
	if runtime.GOOS == "windows" {
		dir := filepath.Dir(path)
		base := filepath.Base(path)

		if target, err := os.Readlink(dir); err == nil {
			dir = target
		}

		resolvedPath := filepath.Join(dir, base)

		if rp, err := filepath.EvalSymlinks(resolvedPath); err == nil {
			return rp
		}
		return resolvedPath
	}
	return path
}
