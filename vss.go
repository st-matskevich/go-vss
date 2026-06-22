//go:build !windows
// +build !windows

package vss

import (
	"errors"
)

type Snapshotter struct{}

func (v *Snapshotter) CreateSnapshot(drive string, timeout int, opts ...Option) (*Snapshot, error) {
	return nil, errors.ErrUnsupported
}

func (v *Snapshotter) Release() error {
	return errors.ErrUnsupported
}
