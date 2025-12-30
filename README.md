# which

Cross-platform `which` command for locating executables in PATH.

## Build

```
go build
```

## Usage

```
which <program>
```

Prints the full path to the executable if found in PATH. Returns exit code 1 if not found.

### Examples

```
$ which go
/usr/local/go/bin/go

$ which notepad
C:\Windows\System32\notepad.exe
```

## Notes

- On Windows, automatically searches for files with PATHEXT extensions (.exe, .bat, .cmd, etc.)
- On Windows, also checks the current directory
- On Unix, checks execute permissions

## License

GPL-2.0
