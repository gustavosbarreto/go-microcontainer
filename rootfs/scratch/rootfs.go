package scratch

import (
	"os"
	"path"
)

type scratchRootFS struct {
}

func NewRootFS() *scratchRootFS {
	return &scratchRootFS{}
}

func (fs *scratchRootFS) PullRootFS() error {
	if err := os.MkdirAll(path.Join("rootfs", "proc"), os.FileMode(0700)); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join("rootfs", "tmp"), os.FileMode(0700)); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join("rootfs", "mnt"), os.FileMode(0700)); err != nil {
		return err
	}

	return nil
}
