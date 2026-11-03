package updater

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
)

func cleanupOld(exe string) {
	if exe == "" {
		return
	}
	os.Remove(exe + ".old")
}

func swapBinary(src, dst string) error {
	return swapBinaryFor(runtime.GOOS, src, dst)
}

func swapBinaryFor(goos, src, dst string) error {
	if goos != "windows" {
		err := os.Rename(src, dst)
		if err != nil {
			os.Remove(src)
			if errors.Is(err, fs.ErrPermission) {
				return fmt.Errorf("cannot replace %s: permission denied (try re-running with sudo / as Administrator): %w", dst, err)
			}
			return err
		}
		return nil
	}

	// Windows: old := dst + ".old"
	old := dst + ".old"
	os.Remove(old)

	err := os.Rename(dst, old)
	if err != nil {
		if _, statErr := os.Stat(dst); statErr == nil {
			os.Remove(src)
			if errors.Is(err, fs.ErrPermission) {
				return fmt.Errorf("cannot replace %s: permission denied (try re-running with sudo / as Administrator): %w", dst, err)
			}
			return err
		}
	}

	err = os.Rename(src, dst)
	if err != nil {
		os.Rename(old, dst)
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("cannot replace %s: permission denied (try re-running with sudo / as Administrator): %w", dst, err)
		}
		return err
	}

	return nil
}
