package main

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"

	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
)

var (
	disallowedBytesRx = regexp.MustCompile("[^\t\r\n\x20-\x7f]")
	ignoredDirRx      = regexp.MustCompile(`^(\.git|/__pycache__|bin|dist)|\.?venv$`)
	ignoredFilenameRx = regexp.MustCompile(`\.(ai|ico|md(?:\.(?:tmpl|yaml))?|pdf|png|syso)$`)
)

func run() error {
	var lintErrs []error
	if err := fs.WalkDir(os.DirFS("."), ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		switch dirEntry.Type() & fs.ModeType {
		case 0:
			if disallowedBytesRx.Match([]byte(path)) {
				lintErrs = append(lintErrs, fmt.Errorf("%q: disallowed filename", path))
			}
			if ignoredFilenameRx.Match([]byte(path)) {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if disallowedBytesRx.Match(data) {
				lintErrs = append(lintErrs, fmt.Errorf("%q: disallowed file contents", path))
			}
		case fs.ModeDir:
			if ignoredDirRx.Match([]byte(path)) {
				return fs.SkipDir
			}
			return nil
		default:
			lintErrs = append(lintErrs, fmt.Errorf("%q: disallowed entry type", path))
		}
		return nil
	}); err != nil {
		return err
	}
	return chezmoierrors.Combine(lintErrs...)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
