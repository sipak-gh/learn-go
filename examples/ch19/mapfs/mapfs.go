// Package mapfs finds files in any file system.
package mapfs

import (
	"io/fs"
	"path"
)

// FindExt returns the paths of the regular files in fsys whose names
// end in ext, such as ".go", in lexical order.
func FindExt(fsys fs.FS, ext string) ([]string, error) {
	var found []string
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() && path.Ext(p) == ext {
			found = append(found, p)
		}
		return nil
	})
	return found, err
}
