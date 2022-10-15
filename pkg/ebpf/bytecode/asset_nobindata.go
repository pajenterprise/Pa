// +build !ebpf_bindata

package bytecode

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

// GetReader returns a new AssetReader for the specified file asset
func GetReader(dir, name string) (AssetReader, error) {
	fullPath, err := filepath.Abs(path.Join(dir, path.Base(name)))
	if err != nil {
		return nil, err
	}

	asset, err := os.Open(fullPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find asset %s", fullPath)
	}

	return asset, nil
}
