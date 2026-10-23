//go:build windows
// +build windows

package vss

import (
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// NOTE: Microsoft Documentation can be found here: https://learn.microsoft.com/en-us/windows/win32/api/vss/nn-vss-ivssenumobject
type IVssEnumObject struct {
	ole.IUnknown
}

// VTable for IVssEnumObject
type IVssEnumObjectVTable struct {
	ole.IUnknownVtbl
	next uintptr
}

// The VssObjectProperties structure holds one shadow copy returned by IVssBackupComponents::Query.
type VssObjectProperties struct {
	objectType uint32
	_          uint32 // Windows aligns the union below to 8 bytes
	snapshot   VssSnapshotProperties
	_          uint32 // On 386 the union is 4 bytes longer than VssSnapshotProperties is in Go
}

// Returns pointer to IVssEnumObjectVTable
func (enum *IVssEnumObject) getVTable() *IVssEnumObjectVTable {
	return (*IVssEnumObjectVTable)(unsafe.Pointer(enum.RawVTable))
}

// The Next method reads the next object into object, or returns false when there are no more objects.
func (enum *IVssEnumObject) Next(object *VssObjectProperties) (bool, error) {
	var fetched uint32
	code, _, _ := syscall.Syscall6(enum.getVTable().next, 4, uintptr(unsafe.Pointer(enum)), 1, uintptr(unsafe.Pointer(object)), uintptr(unsafe.Pointer(&fetched)), 0, 0)
	if HRESULT(code) == S_FALSE {
		return false, nil
	}
	return fetched == 1, CreateVSSError("IVssEnumObject.Next", code)
}

// Frees the strings that VSS allocated for the snapshot properties
func freeSnapshotProperties(properties *VssSnapshotProperties) {
	syscall.NewLazyDLL("VssApi.dll").NewProc("VssFreeSnapshotProperties").Call(uintptr(unsafe.Pointer(properties)))
}
