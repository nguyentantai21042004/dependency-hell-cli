package scanner

import (
	"io/fs"
	"path/filepath"
)

// CalculateDirSize calculates the total size of a directory
func CalculateDirSize(path string) (int64, error) {
	expandedPath := ExpandHome(path)

	if !PathExists(expandedPath) {
		return 0, nil
	}

	var size int64
	err := filepath.WalkDir(expandedPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't access
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return size, nil
}

// ScanMultiplePaths scans multiple paths and returns total size
func ScanMultiplePaths(paths []string) (int64, error) {
	var total int64

	for _, path := range paths {
		size, err := CalculateDirSize(path)
		if err != nil {
			// Continue on error, just skip this path
			continue
		}
		total += size
	}

	return total, nil
}

// CalculatePathSizes calculates sizes for multiple paths individually
func CalculatePathSizes(paths map[string]string) map[string]int64 {
	sizes := make(map[string]int64)

	for description, path := range paths {
		size, err := CalculateDirSize(path)
		if err != nil {
			sizes[description] = 0
			continue
		}
		sizes[description] = size
	}

	return sizes
}
