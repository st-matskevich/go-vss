package vss

import (
	"errors"

	ole "github.com/go-ole/go-ole"
)

// WriterMetadata identifies a writer that reported metadata for the snapshot.
type WriterMetadata struct {
	ClassID ole.GUID // The writer class, the same for all its instances
	Name    string

	readComponents func() ([]WriterComponent, error)
}

// Components reads the writer's components and their files. It works only
// while the callback given to WithWriterMetadata runs.
func (w WriterMetadata) Components() ([]WriterComponent, error) {
	if w.readComponents == nil {
		return nil, errors.New("writer metadata is not available")
	}
	return w.readComponents()
}

// WriterComponent is a part of a writer's data, such as a database.
type WriterComponent struct {
	LogicalPath string
	Name        string
	// Files of a file group component
	Files []FileSet
	// Files of a database component
	DatabaseFiles    []FileSet
	DatabaseLogFiles []FileSet
}

// FileSet is the files in Path that match Filespec, and also those in the
// subdirectories of Path when Recursive is set.
type FileSet struct {
	Path      string
	Filespec  string
	Recursive bool
}
