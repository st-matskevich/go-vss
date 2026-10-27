//go:build windows
// +build windows

package vss

import (
	"fmt"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// NOTE: Microsoft Documentation can be found here: https://learn.microsoft.com/en-us/windows/win32/api/vsbackup/nl-vsbackup-ivssexaminewritermetadata
// The IVssExamineWriterMetadata interface reads the metadata that a writer reported to GatherWriterMetadata.
type IVssExamineWriterMetadata struct {
	ole.IUnknown
}

// VTable for IVssExamineWriterMetadata. Commented members are simply not used till now.
type IVssExamineWriterMetadataVTable struct {
	ole.IUnknownVtbl
	getIdentity    uintptr
	getFileCounts  uintptr
	getIncludeFile uintptr
	getExcludeFile uintptr
	getComponent   uintptr
	// getRestoreMethod            uintptr
	// getAlternateLocationMapping uintptr
	// getBackupSchema             uintptr
	// getDocument                 uintptr
	// saveAsXML                   uintptr
	// loadFromXML                 uintptr
}

// Returns pointer to IVssExamineWriterMetadataVTable
func (m *IVssExamineWriterMetadata) getVTable() *IVssExamineWriterMetadataVTable {
	return (*IVssExamineWriterMetadataVTable)(unsafe.Pointer(m.RawVTable))
}

// The GetIdentity method returns the writer's instance ID, class ID and name.
func (m *IVssExamineWriterMetadata) GetIdentity() (instanceID, classID ole.GUID, name string, err error) {
	var bstrName *uint16
	var usage, source uint32
	code, _, _ := syscall.Syscall6(m.getVTable().getIdentity, 6, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(&instanceID)), uintptr(unsafe.Pointer(&classID)), uintptr(unsafe.Pointer(&bstrName)), uintptr(unsafe.Pointer(&usage)), uintptr(unsafe.Pointer(&source)))
	if err = CreateVSSError("IVssExamineWriterMetadata.GetIdentity", code); err != nil {
		return instanceID, classID, "", err
	}
	return instanceID, classID, takeBSTR(bstrName), nil
}

// The GetFileCounts method returns the number of include files, exclude files and components of the writer.
func (m *IVssExamineWriterMetadata) GetFileCounts() (includeFiles, excludeFiles, components uint32, err error) {
	code, _, _ := syscall.Syscall6(m.getVTable().getFileCounts, 4, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(&includeFiles)), uintptr(unsafe.Pointer(&excludeFiles)), uintptr(unsafe.Pointer(&components)), 0, 0)
	return includeFiles, excludeFiles, components, CreateVSSError("IVssExamineWriterMetadata.GetFileCounts", code)
}

// The GetComponent method returns a component of the writer. The caller must release it.
func (m *IVssExamineWriterMetadata) GetComponent(index uint32) (*IVssWMComponent, error) {
	var component *IVssWMComponent
	code, _, _ := syscall.Syscall(m.getVTable().getComponent, 3, uintptr(unsafe.Pointer(m)), uintptr(index), uintptr(unsafe.Pointer(&component)))
	if err := CreateVSSError("IVssExamineWriterMetadata.GetComponent", code); err != nil {
		return nil, err
	}
	if component == nil {
		return nil, fmt.Errorf("IVssExamineWriterMetadata.GetComponent failed to return a valid IVssWMComponent object")
	}
	return component, nil
}
