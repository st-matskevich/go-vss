//go:build windows
// +build windows

package vss

import (
	"fmt"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// NOTE: Microsoft Documentation can be found here: https://learn.microsoft.com/en-us/windows/win32/api/vsbackup/nl-vsbackup-ivsswmcomponent
// The IVssWMComponent interface reads one component of a writer's metadata.
type IVssWMComponent struct {
	ole.IUnknown
}

// VTable for IVssWMComponent. Commented members are simply not used till now.
type IVssWMComponentVTable struct {
	ole.IUnknownVtbl
	getComponentInfo   uintptr
	freeComponentInfo  uintptr
	getFile            uintptr
	getDatabaseFile    uintptr
	getDatabaseLogFile uintptr
	// getDependency      uintptr
}

// NOTE: Microsoft Documentation can be found here: https://learn.microsoft.com/en-us/windows/win32/api/vsbackup/ns-vsbackup-vss_componentinfo
// vssComponentInfo has the layout of the VSS_COMPONENTINFO structure that GetComponentInfo allocates.
type vssComponentInfo struct {
	componentType          int32
	logicalPath            *uint16
	componentName          *uint16
	caption                *uint16
	icon                   *byte
	iconSize               uint32
	restoreMetadata        bool
	notifyOnBackupComplete bool
	selectable             bool
	selectableForRestore   bool
	componentFlags         uint32
	fileCount              uint32
	databaseFileCount      uint32
	logFileCount           uint32
	dependencyCount        uint32
}

// ComponentInfo is the name and the file counts of a component.
type ComponentInfo struct {
	LogicalPath       string
	Name              string
	FileCount         uint32 // Files of a file group component
	DatabaseFileCount uint32 // Database files of a database component
	LogFileCount      uint32 // Log files of a database component
}

// Returns pointer to IVssWMComponentVTable
func (c *IVssWMComponent) getVTable() *IVssWMComponentVTable {
	return (*IVssWMComponentVTable)(unsafe.Pointer(c.RawVTable))
}

// The GetComponentInfo method returns the component's name and file counts.
func (c *IVssWMComponent) GetComponentInfo() (ComponentInfo, error) {
	var info *vssComponentInfo
	code, _, _ := syscall.Syscall(c.getVTable().getComponentInfo, 2, uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(&info)), 0)
	if err := CreateVSSError("IVssWMComponent.GetComponentInfo", code); err != nil {
		return ComponentInfo{}, err
	}
	if info == nil {
		return ComponentInfo{}, fmt.Errorf("IVssWMComponent.GetComponentInfo failed to return a valid VSS_COMPONENTINFO structure")
	}
	// VSS allocated the structure, it is freed with FreeComponentInfo once copied
	defer syscall.Syscall(c.getVTable().freeComponentInfo, 2, uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(info)), 0)

	return ComponentInfo{
		LogicalPath:       ole.BstrToString(info.logicalPath),
		Name:              ole.BstrToString(info.componentName),
		FileCount:         info.fileCount,
		DatabaseFileCount: info.databaseFileCount,
		LogFileCount:      info.logFileCount,
	}, nil
}

// The GetFile method returns a file set of a file group component. The caller must release it.
func (c *IVssWMComponent) GetFile(index uint32) (*IVssWMFiledesc, error) {
	return c.getFiledesc(c.getVTable().getFile, "IVssWMComponent.GetFile", index)
}

// The GetDatabaseFile method returns a database file set of a database component. The caller must release it.
func (c *IVssWMComponent) GetDatabaseFile(index uint32) (*IVssWMFiledesc, error) {
	return c.getFiledesc(c.getVTable().getDatabaseFile, "IVssWMComponent.GetDatabaseFile", index)
}

// The GetDatabaseLogFile method returns a log file set of a database component. The caller must release it.
func (c *IVssWMComponent) GetDatabaseLogFile(index uint32) (*IVssWMFiledesc, error) {
	return c.getFiledesc(c.getVTable().getDatabaseLogFile, "IVssWMComponent.GetDatabaseLogFile", index)
}

func (c *IVssWMComponent) getFiledesc(method uintptr, name string, index uint32) (*IVssWMFiledesc, error) {
	var filedesc *IVssWMFiledesc
	code, _, _ := syscall.Syscall(method, 3, uintptr(unsafe.Pointer(c)), uintptr(index), uintptr(unsafe.Pointer(&filedesc)))
	if err := CreateVSSError(name, code); err != nil {
		return nil, err
	}
	if filedesc == nil {
		return nil, fmt.Errorf("%s failed to return a valid IVssWMFiledesc object", name)
	}
	return filedesc, nil
}
