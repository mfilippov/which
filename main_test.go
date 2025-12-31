package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsExecutable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	t.Run("non-existent file returns false", func(t *testing.T) {
		if isExecutable(filepath.Join(tmpDir, "nonexistent")) {
			t.Error("Expected false for non-existent file")
		}
	})

	t.Run("directory returns false", func(t *testing.T) {
		if isExecutable(tmpDir) {
			t.Error("Expected false for directory")
		}
	})

	t.Run("regular file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "testfile")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if runtime.GOOS == "windows" {
			if !isExecutable(testFile) {
				t.Error("Expected true for regular file on Windows")
			}
		} else {
			if isExecutable(testFile) {
				t.Error("Expected false for file without execute permission")
			}

			if err := os.Chmod(testFile, 0755); err != nil {
				t.Fatalf("Failed to chmod: %v", err)
			}
			if !isExecutable(testFile) {
				t.Error("Expected true for file with execute permission")
			}
		}
	})
}

func TestFindInDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if runtime.GOOS == "windows" {
		if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
			tmpDir = resolved
		}
	}

	if runtime.GOOS == "windows" {
		t.Run("finds .exe file", func(t *testing.T) {
			exeFile := filepath.Join(tmpDir, "testprog.exe")
			if err := os.WriteFile(exeFile, []byte("test"), 0755); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := findInDir(tmpDir, "testprog")
			if !strings.EqualFold(result, exeFile) {
				t.Errorf("Expected %s, got %s", exeFile, result)
			}
		})

		t.Run("finds .bat file", func(t *testing.T) {
			batFile := filepath.Join(tmpDir, "script.bat")
			if err := os.WriteFile(batFile, []byte("test"), 0755); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := findInDir(tmpDir, "script")
			if !strings.EqualFold(result, batFile) {
				t.Errorf("Expected %s, got %s", batFile, result)
			}
		})

		t.Run("finds .cmd file", func(t *testing.T) {
			cmdFile := filepath.Join(tmpDir, "script2.cmd")
			if err := os.WriteFile(cmdFile, []byte("test"), 0755); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := findInDir(tmpDir, "script2")
			if !strings.EqualFold(result, cmdFile) {
				t.Errorf("Expected %s, got %s", cmdFile, result)
			}
		})

		t.Run("prefers .exe over .bat", func(t *testing.T) {
			exeFile := filepath.Join(tmpDir, "both.exe")
			batFile := filepath.Join(tmpDir, "both.bat")
			if err := os.WriteFile(exeFile, []byte("test"), 0755); err != nil {
				t.Fatalf("Failed to create exe file: %v", err)
			}
			if err := os.WriteFile(batFile, []byte("test"), 0755); err != nil {
				t.Fatalf("Failed to create bat file: %v", err)
			}

			result := findInDir(tmpDir, "both")
			if !strings.EqualFold(result, exeFile) {
				t.Errorf("Expected %s (exe preferred), got %s", exeFile, result)
			}
		})

		t.Run("finds file with explicit extension", func(t *testing.T) {
			batFile := filepath.Join(tmpDir, "explicit.bat")
			if err := os.WriteFile(batFile, []byte("test"), 0755); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := findInDir(tmpDir, "explicit.bat")
			if !strings.EqualFold(result, batFile) {
				t.Errorf("Expected %s, got %s", batFile, result)
			}
		})

		t.Run("explicit extension not found returns empty", func(t *testing.T) {
			result := findInDir(tmpDir, "nonexistent.exe")
			if result != "" {
				t.Errorf("Expected empty string, got %s", result)
			}
		})
	} else {
		t.Run("finds executable file on Unix", func(t *testing.T) {
			exeFile := filepath.Join(tmpDir, "unixprog")
			if err := os.WriteFile(exeFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := findInDir(tmpDir, "unixprog")
			if !strings.EqualFold(result, exeFile) {
				t.Errorf("Expected %s, got %s", exeFile, result)
			}
		})

		t.Run("non-executable file not found on Unix", func(t *testing.T) {
			nonExeFile := filepath.Join(tmpDir, "nonexe")
			if err := os.WriteFile(nonExeFile, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := findInDir(tmpDir, "nonexe")
			if result != "" {
				t.Errorf("Expected empty string for non-executable, got %s", result)
			}
		})
	}

	t.Run("not found returns empty string", func(t *testing.T) {
		result := findInDir(tmpDir, "doesnotexist")
		if result != "" {
			t.Errorf("Expected empty string, got %s", result)
		}
	})
}

func TestFindExecutable(t *testing.T) {
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", originalPath) })

	tmpDir1, err := os.MkdirTemp("", "which-test1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir1) })

	tmpDir2, err := os.MkdirTemp("", "which-test2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir2) })

	if runtime.GOOS == "windows" {
		if resolved, err := filepath.EvalSymlinks(tmpDir1); err == nil {
			tmpDir1 = resolved
		}
		if resolved, err := filepath.EvalSymlinks(tmpDir2); err == nil {
			tmpDir2 = resolved
		}
	}

	var testExe1, testExe2 string
	if runtime.GOOS == "windows" {
		testExe1 = filepath.Join(tmpDir1, "prog1.exe")
		testExe2 = filepath.Join(tmpDir2, "prog2.exe")
	} else {
		testExe1 = filepath.Join(tmpDir1, "prog1")
		testExe2 = filepath.Join(tmpDir2, "prog2")
	}

	if err := os.WriteFile(testExe1, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testExe2, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	pathSep := ":"
	if runtime.GOOS == "windows" {
		pathSep = ";"
	}
	if err := os.Setenv("PATH", tmpDir1+pathSep+tmpDir2); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	t.Run("finds program in first PATH directory", func(t *testing.T) {
		result := findExecutable("prog1")
		if !strings.EqualFold(result, testExe1) {
			t.Errorf("Expected %s, got %s", testExe1, result)
		}
	})

	t.Run("finds program in second PATH directory", func(t *testing.T) {
		result := findExecutable("prog2")
		if !strings.EqualFold(result, testExe2) {
			t.Errorf("Expected %s, got %s", testExe2, result)
		}
	})

	t.Run("returns first match when in multiple directories", func(t *testing.T) {
		var dupExe string
		if runtime.GOOS == "windows" {
			dupExe = filepath.Join(tmpDir2, "prog1.exe")
		} else {
			dupExe = filepath.Join(tmpDir2, "prog1")
		}
		if err := os.WriteFile(dupExe, []byte("test2"), 0755); err != nil {
			t.Fatalf("Failed to create duplicate file: %v", err)
		}

		result := findExecutable("prog1")
		if !strings.EqualFold(result, testExe1) {
			t.Errorf("Expected first match %s, got %s", testExe1, result)
		}
	})

	t.Run("not found returns empty string", func(t *testing.T) {
		result := findExecutable("nonexistent")
		if result != "" {
			t.Errorf("Expected empty string, got %s", result)
		}
	})

	t.Run("empty PATH returns empty string", func(t *testing.T) {
		if err := os.Setenv("PATH", ""); err != nil {
			t.Fatalf("Failed to set PATH: %v", err)
		}
		result := findExecutable("prog1")
		if result != "" {
			t.Errorf("Expected empty string for empty PATH, got %s", result)
		}
	})
}

func TestFindExecutableWithEmptyDirs(t *testing.T) {
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", originalPath) })

	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if runtime.GOOS == "windows" {
		if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
			tmpDir = resolved
		}
	}

	var testExe string
	if runtime.GOOS == "windows" {
		testExe = filepath.Join(tmpDir, "prog.exe")
	} else {
		testExe = filepath.Join(tmpDir, "prog")
	}
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	pathSep := ":"
	if runtime.GOOS == "windows" {
		pathSep = ";"
	}

	if err := os.Setenv("PATH", ""+pathSep+tmpDir+pathSep+""); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	if runtime.GOOS == "windows" {
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get cwd: %v", err)
		}
		if err := os.Chdir(os.TempDir()); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		t.Cleanup(func() { _ = os.Chdir(origDir) })
	}

	result := findExecutable("prog")
	if !strings.EqualFold(result, testExe) {
		t.Errorf("Expected %s, got %s", testExe, result)
	}
}

func TestGetExtensions(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Run("returns nil on non-Windows", func(t *testing.T) {
			exts := getExtensions()
			if exts != nil {
				t.Errorf("Expected nil on non-Windows, got %v", exts)
			}
		})
		return
	}

	originalPathExt := os.Getenv("PATHEXT")
	t.Cleanup(func() { _ = os.Setenv("PATHEXT", originalPathExt) })

	t.Run("returns default extensions when PATHEXT is empty", func(t *testing.T) {
		if err := os.Setenv("PATHEXT", ""); err != nil {
			t.Fatalf("Failed to set PATHEXT: %v", err)
		}
		exts := getExtensions()
		expected := []string{".COM", ".EXE", ".BAT", ".CMD"}
		if len(exts) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, exts)
		}
	})

	t.Run("parses PATHEXT correctly", func(t *testing.T) {
		if err := os.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD;.PS1"); err != nil {
			t.Fatalf("Failed to set PATHEXT: %v", err)
		}
		exts := getExtensions()
		if len(exts) != 5 {
			t.Errorf("Expected 5 extensions, got %d: %v", len(exts), exts)
		}
		if exts[4] != ".PS1" {
			t.Errorf("Expected .PS1 as last extension, got %s", exts[4])
		}
	})

	t.Run("handles empty entries in PATHEXT", func(t *testing.T) {
		if err := os.Setenv("PATHEXT", ".EXE;;.BAT"); err != nil {
			t.Fatalf("Failed to set PATHEXT: %v", err)
		}
		exts := getExtensions()
		if len(exts) != 2 {
			t.Errorf("Expected 2 extensions, got %d: %v", len(exts), exts)
		}
	})
}

func TestIsPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple name", "program", false},
		{"with forward slash", "./program", true},
		{"with backslash", ".\\program", true},
		{"absolute unix path", "/usr/bin/program", true},
		{"absolute windows path", "C:\\Windows\\program", true},
		{"relative path", "subdir/program", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPath(tt.input)
			if result != tt.expected {
				t.Errorf("isPath(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindExecutableWithPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if runtime.GOOS == "windows" {
		if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
			tmpDir = resolved
		}
	}

	var testExe string
	if runtime.GOOS == "windows" {
		testExe = filepath.Join(tmpDir, "myprog.exe")
	} else {
		testExe = filepath.Join(tmpDir, "myprog")
	}
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("finds file with explicit path", func(t *testing.T) {
		result := findExecutable(testExe)
		if !strings.EqualFold(result, testExe) {
			t.Errorf("Expected %s, got %s", testExe, result)
		}
	})

	t.Run("returns empty for non-existent path", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "nonexistent")
		if runtime.GOOS == "windows" {
			nonExistent += ".exe"
		}
		result := findExecutable(nonExistent)
		if result != "" {
			t.Errorf("Expected empty string, got %s", result)
		}
	})
}

func TestFindInCurrentDirectory(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Current directory search is Windows-specific")
	}

	originalPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", originalPath) })

	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolved
	}

	testExe := filepath.Join(tmpDir, "cwdprog.exe")
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if err := os.Setenv("PATH", origDir); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	t.Run("finds executable in current directory on Windows", func(t *testing.T) {
		result := findExecutable("cwdprog")
		if !strings.EqualFold(result, testExe) {
			t.Errorf("Expected %s, got %s", testExe, result)
		}
	})
}

func TestCaseInsensitiveExtension(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Extension handling is Windows-specific")
	}

	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	testExe := filepath.Join(tmpDir, "caseprog.EXE")
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("finds file with different case extension", func(t *testing.T) {
		result := findInDir(tmpDir, "caseprog.exe")
		if result == "" {
			t.Error("Expected to find file with case-insensitive extension match")
		}
		if !strings.EqualFold(filepath.Base(result), "caseprog.exe") {
			t.Errorf("Unexpected result: %s", result)
		}
	})
}

func TestCaseSensitiveFilesystem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows filesystem is always case-insensitive")
	}

	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	lowerFile := filepath.Join(tmpDir, "prog")
	upperFile := filepath.Join(tmpDir, "PROG")

	if err := os.WriteFile(lowerFile, []byte("lower"), 0755); err != nil {
		t.Fatalf("Failed to create lower file: %v", err)
	}

	err = os.WriteFile(upperFile, []byte("upper"), 0755)
	if err != nil {
		t.Skip("Filesystem is case-insensitive, skipping test")
	}

	lowerInfo, _ := os.Stat(lowerFile)
	upperInfo, _ := os.Stat(upperFile)
	if os.SameFile(lowerInfo, upperInfo) {
		t.Skip("Filesystem is case-insensitive, skipping test")
	}

	t.Run("finds exact case match on case-sensitive filesystem", func(t *testing.T) {
		result := findInDir(tmpDir, "prog")
		if result != lowerFile {
			t.Errorf("Expected %s, got %s", lowerFile, result)
		}
	})

	t.Run("finds uppercase file when searching uppercase", func(t *testing.T) {
		result := findInDir(tmpDir, "PROG")
		if result != upperFile {
			t.Errorf("Expected %s, got %s", upperFile, result)
		}
	})
}

func TestNormalizePath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("normalizePath is Windows-specific")
	}

	tmpDir, err := os.MkdirTemp("", "which-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolved
	}

	t.Run("normalizes extension case", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.exe")
		if err := os.WriteFile(testFile, []byte("test"), 0755); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		inputPath := filepath.Join(tmpDir, "test.EXE")
		result := normalizePath(inputPath)

		if !strings.HasSuffix(result, "test.exe") {
			t.Errorf("Expected path ending with 'test.exe', got %s", result)
		}
	})
}

func TestJunctionResolution(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Junction points are Windows-specific")
	}

	tmpDir, err := os.MkdirTemp("", "which-junction-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolved
	}

	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	testExe := filepath.Join(targetDir, "prog.exe")
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	junctionDir := filepath.Join(tmpDir, "junction")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", junctionDir, targetDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create junction: %v", err)
	}

	t.Run("finds executable through junction", func(t *testing.T) {
		result := findInDir(junctionDir, "prog")
		if result == "" {
			t.Error("Expected to find executable through junction")
		}
	})

	t.Run("normalizes case through junction", func(t *testing.T) {
		inputPath := filepath.Join(junctionDir, "prog.EXE")
		result := normalizePath(inputPath)

		if !strings.HasSuffix(result, "prog.exe") {
			t.Errorf("Expected path ending with 'prog.exe', got %s", result)
		}
	})

	t.Run("resolves junction to target", func(t *testing.T) {
		inputPath := filepath.Join(junctionDir, "prog.EXE")
		result := normalizePath(inputPath)

		if !strings.Contains(result, "target") {
			t.Errorf("Expected path to contain 'target' (resolved junction), got %s", result)
		}
	})
}

func TestFindExecutableThroughJunction(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Junction points are Windows-specific")
	}

	originalPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", originalPath) })

	tmpDir, err := os.MkdirTemp("", "which-junction-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolved
	}

	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	testExe := filepath.Join(targetDir, "junctionprog.exe")
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	junctionDir := filepath.Join(tmpDir, "junction")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", junctionDir, targetDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create junction: %v", err)
	}

	if err := os.Setenv("PATH", junctionDir); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	t.Run("finds and normalizes executable through junction in PATH", func(t *testing.T) {
		result := findExecutable("junctionprog")
		if result == "" {
			t.Fatal("Expected to find executable")
		}

		if strings.HasSuffix(result, ".EXE") {
			t.Errorf("Expected lowercase extension, got uppercase: %s", result)
		}
	})
}
