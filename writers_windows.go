//go:build windows
// +build windows

package vss

import (
	"errors"
	"fmt"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// inspectWriterMetadata passes the metadata gathered by GatherWriterMetadata to inspect.
func inspectWriterMetadata(components *IVssBackupComponents, inspect func([]WriterMetadata) error) error {
	count, err := components.GetWriterMetadataCount()
	if err != nil {
		return fmt.Errorf("VSS_METADATA - Shadow copy creation failed: GetWriterMetadataCount, err: %w", err)
	}

	// The metadata is released when inspect returns, so it can't be read after that
	inspecting := true
	var examined []*IVssExamineWriterMetadata
	defer func() {
		inspecting = false
		for _, metadata := range examined {
			metadata.Release()
		}
	}()

	writers := make([]WriterMetadata, 0, count)
	for i := uint32(0); i < count; i++ {
		metadata, err := components.GetWriterMetadata(i)
		if err != nil {
			return fmt.Errorf("VSS_METADATA - Shadow copy creation failed: GetWriterMetadata, err: %w", err)
		}
		examined = append(examined, metadata)

		_, classID, name, err := metadata.GetIdentity()
		if err != nil {
			return fmt.Errorf("VSS_METADATA - Shadow copy creation failed: GetIdentity, err: %w", err)
		}

		writers = append(writers, WriterMetadata{
			ClassID: classID,
			Name:    name,
			readComponents: func() ([]WriterComponent, error) {
				if !inspecting {
					return nil, errors.New("writer metadata can be read only while the WithWriterMetadata callback runs")
				}
				return readComponents(metadata)
			},
		})
	}

	if err := inspect(writers); err != nil {
		return fmt.Errorf("VSS_METADATA - Shadow copy creation stopped by the writer metadata callback, err: %w", err)
	}
	return nil
}

// readComponents reads all components of a writer with their files.
func readComponents(metadata *IVssExamineWriterMetadata) ([]WriterComponent, error) {
	_, _, count, err := metadata.GetFileCounts()
	if err != nil {
		return nil, err
	}

	components := make([]WriterComponent, 0, count)
	for i := uint32(0); i < count; i++ {
		component, err := readComponent(metadata, i)
		if err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	return components, nil
}

func readComponent(metadata *IVssExamineWriterMetadata, index uint32) (WriterComponent, error) {
	component, err := metadata.GetComponent(index)
	if err != nil {
		return WriterComponent{}, err
	}
	defer component.Release()

	info, err := component.GetComponentInfo()
	if err != nil {
		return WriterComponent{}, err
	}

	result := WriterComponent{LogicalPath: info.LogicalPath, Name: info.Name}
	if result.Files, err = readFileSets(info.FileCount, component.GetFile); err != nil {
		return WriterComponent{}, err
	}
	if result.DatabaseFiles, err = readFileSets(info.DatabaseFileCount, component.GetDatabaseFile); err != nil {
		return WriterComponent{}, err
	}
	if result.DatabaseLogFiles, err = readFileSets(info.LogFileCount, component.GetDatabaseLogFile); err != nil {
		return WriterComponent{}, err
	}
	return result, nil
}

// readFileSets reads count file sets with get, one of the GetFile methods of IVssWMComponent.
func readFileSets(count uint32, get func(uint32) (*IVssWMFiledesc, error)) ([]FileSet, error) {
	fileSets := make([]FileSet, 0, count)
	for i := uint32(0); i < count; i++ {
		fileSet, err := readFileSet(get, i)
		if err != nil {
			return nil, err
		}
		fileSets = append(fileSets, fileSet)
	}
	return fileSets, nil
}

func readFileSet(get func(uint32) (*IVssWMFiledesc, error), index uint32) (FileSet, error) {
	filedesc, err := get(index)
	if err != nil {
		return FileSet{}, err
	}
	defer filedesc.Release()

	path, err := filedesc.GetPath()
	if err != nil {
		return FileSet{}, err
	}
	filespec, err := filedesc.GetFilespec()
	if err != nil {
		return FileSet{}, err
	}
	recursive, err := filedesc.GetRecursive()
	if err != nil {
		return FileSet{}, err
	}
	return FileSet{Path: path, Filespec: filespec, Recursive: recursive}, nil
}

// takeBSTR copies a BSTR that VSS handed over to the caller and frees it.
func takeBSTR(bstr *uint16) string {
	if bstr == nil {
		return ""
	}
	defer ole.SysFreeString((*int16)(unsafe.Pointer(bstr)))
	return ole.BstrToString(bstr)
}
