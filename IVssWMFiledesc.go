//go:build windows
// +build windows

package vss

import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// NOTE: Microsoft Documentation can be found here: https://learn.microsoft.com/en-us/windows/win32/api/vswriter/nl-vswriter-ivsswmfiledesc
// The IVssWMFiledesc interface describes a set of files in a component of a writer's metadata.
type IVssWMFiledesc struct {
	ole.IUnknown
}

// VTable for IVssWMFiledesc. Commented members are simply not used till now.
type IVssWMFiledescVTable struct {
	ole.IUnknownVtbl
	getPath      uintptr
	getFilespec  uintptr
	getRecursive uintptr
	// getAlternateLocation uintptr
	// getBackupTypeMask    uintptr
}

// Returns pointer to IVssWMFiledescVTable
func (f *IVssWMFiledesc) getVTable() *IVssWMFiledescVTable {
	return (*IVssWMFiledescVTable)(unsafe.Pointer(f.RawVTable))
}

// The GetPath method returns the directory that holds the files.
func (f *IVssWMFiledesc) GetPath() (string, error) {
	return f.getString(f.getVTable().getPath, "IVssWMFiledesc.GetPath")
}

// The GetFilespec method returns the file specification, which can contain wildcards.
func (f *IVssWMFiledesc) GetFilespec() (string, error) {
	return f.getString(f.getVTable().getFilespec, "IVssWMFiledesc.GetFilespec")
}

// The GetRecursive method reports whether the files in the subdirectories of the path are included.
func (f *IVssWMFiledesc) GetRecursive() (bool, error) {
	// A C++ bool, one byte
	var recursive uint8
	code, _, _ := syscall.Syscall(f.getVTable().getRecursive, 2, uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&recursive)), 0)
	return recursive != 0, CreateVSSError("IVssWMFiledesc.GetRecursive", code)
}

func (f *IVssWMFiledesc) getString(method uintptr, name string) (string, error) {
	var value *uint16
	code, _, _ := syscall.Syscall(method, 2, uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(&value)), 0)
	if err := CreateVSSError(name, code); err != nil {
		return "", err
	}
	return takeBSTR(value), nil
}
