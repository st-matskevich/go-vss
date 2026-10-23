//go:build windows
// +build windows

package vss

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// NOTE: Microsoft Documentation can be found here: https://docs.microsoft.com/en-us/windows/win32/api/vsbackup/nl-vsbackup-ivssbackupcomponents, I stole most of the comments from that site
const VSS_Create_VSS_BACKUP_COMPONENTS = "?CreateVssBackupComponents@@YAJPEAPEAVIVssBackupComponents@@@Z"
const VSS_Create_VSS_BACKUP_COMPONENTS_386 = "?CreateVssBackupComponents@@YGJPAPAVIVssBackupComponents@@@Z"

// The IVssBackupComponents interface is used by a requester to poll writers about file status and to run backup/restore operations.
type IVssBackupComponents struct {
	ole.IUnknown
}

// VTable for IVssBackupComponents. Commented members are simply not used till now.
type IVssBackupComponentsVTable struct {
	ole.IUnknownVtbl
	getWriterComponentsCount      uintptr
	getWriterComponents           uintptr
	initializeForBackup           uintptr
	setBackupState                uintptr
	initializeForRestore          uintptr
	setRestoreState               uintptr
	gatherWriterMetadata          uintptr
	getWriterMetadataCount        uintptr
	getWriterMetadata             uintptr
	freeWriterMetadata            uintptr
	addComponent                  uintptr
	prepareForBackup              uintptr
	abortBackup                   uintptr
	gatherWriterStatus            uintptr
	getWriterStatusCount          uintptr
	freeWriterStatus              uintptr
	getWriterStatus               uintptr
	setBackupSucceeded            uintptr
	setBackupOptions              uintptr
	setSelectedForRestore         uintptr
	setRestoreOptions             uintptr
	setAdditionalRestores         uintptr
	setPreviousBackupStamp        uintptr
	saveAsXML                     uintptr
	backupComplete                uintptr
	addAlternativeLocationMapping uintptr
	addRestoreSubcomponent        uintptr
	setFileRestoreStatus          uintptr
	addNewTarget                  uintptr
	setRangesFilePath             uintptr
	preRestore                    uintptr
	postRestore                   uintptr
	setContext                    uintptr
	startSnapshotSet              uintptr
	addToSnapshotSet              uintptr
	doSnapshotSet                 uintptr
	deleteSnapshots               uintptr
	importSnapshots               uintptr
	breakSnapshotSet              uintptr
	getSnapshotProperties         uintptr
	query                         uintptr
	isVolumeSupported             uintptr
	// disableWriterClasses          uintptr
	// enableWriterClasses           uintptr
	// disableWriterInstances        uintptr
	// exposeSnapshot                uintptr
	// revertToSnapshot              uintptr
	// queryRevertStatus             uintptr
}

// Loads VssApi.dll and creates VssBackupComponents
func LoadAndInitVSS() (*IVssBackupComponents, error) {
	// Load DLL
	dllVssApi := syscall.NewLazyDLL("VssApi.dll")
	procCreateVssBackupComponents := dllVssApi.NewProc(VSS_Create_VSS_BACKUP_COMPONENTS)
	if runtime.GOARCH == "386" {
		procCreateVssBackupComponents = dllVssApi.NewProc(VSS_Create_VSS_BACKUP_COMPONENTS_386)
	}

	var components *IVssBackupComponents

	if r, _, _ := procCreateVssBackupComponents.Call(uintptr(unsafe.Pointer(&components))); r == (uintptr)(E_ACCESSDENIED) {
		return nil, fmt.Errorf("VSS_CREATE - Only administrators can create shadow copies")
	} else if r != 0 {
		return nil, fmt.Errorf("VSS_CREATE - Failed to create the VSS backup component: %d", r)
	}
	if components == nil {
		return nil, fmt.Errorf("VSS_CREATE - Failed to create the VSS backup component")
	}
	if err := components.InitializeForBackup(); err != nil {
		components.Release()
		return nil, fmt.Errorf("VSS_INIT - Shadow copy creation failed: InitializeForBackup, err: %w", err)
	}
	return components, nil
}

// Returns pointer to IVssBackupComponentsVTable
func (vss *IVssBackupComponents) getVTable() *IVssBackupComponentsVTable {
	return (*IVssBackupComponentsVTable)(unsafe.Pointer(vss.RawVTable))
}

// The AbortBackup method notifies VSS that a backup operation was terminated.
func (vss *IVssBackupComponents) AbortBackup() error {
	code, _, _ := syscall.Syscall(vss.getVTable().abortBackup, 1, uintptr(unsafe.Pointer(vss)), 0, 0)
	return CreateVSSError("IVssBackupComponents.AbortBackup", code)
}

// The InitializeForBackup method initializes the backup components metadata in preparation for backup.
func (vss *IVssBackupComponents) InitializeForBackup() error {
	ret, _, _ := syscall.Syscall(vss.getVTable().initializeForBackup, 2, uintptr(unsafe.Pointer(vss)), 0, 0)
	return CreateVSSError("IVssBackupComponents.InitializeForBackup", ret)
}

// The GatherWriterMetadata method prompts each writer to send the metadata they have collected. The method will generate an Identify event to communicate with writers.
func (vss *IVssBackupComponents) GatherWriterMetadata() (*IVssAsync, error) {
	var async *IVssAsync
	code, _, _ := syscall.Syscall(vss.getVTable().gatherWriterMetadata, 2, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&async)), 0)
	if code != 0 {
		return nil, CreateVSSError("IVssBackupComponents.GatherWriterMetadata", code)
	}
	return async, nil
}

// The FreeWriterMetadata method frees system resources allocated when IVssBackupComponents::GatherWriterMetadata was called.
func (vss *IVssBackupComponents) FreeWriterMetadata() error {
	code, _, _ := syscall.Syscall(vss.getVTable().freeWriterMetadata, 1, uintptr(unsafe.Pointer(vss)), 0, 0)
	return CreateVSSError("IVssBackupComponents.FreeWriterMetadata", code)
}

// The GatherWriterStatus method prompts each writer to send a status message.
func (vss *IVssBackupComponents) GatherWriterStatus() (*IVssAsync, error) {
	var async *IVssAsync
	code, _, _ := syscall.Syscall(vss.getVTable().gatherWriterStatus, 2, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&async)), 0)
	if code != 0 {
		return nil, CreateVSSError("IVssBackupComponents.GatherWriterStatus", code)
	}
	return async, nil
}

// The FreeWriterStatus method frees system resources allocated during the call to IVssBackupComponents::GatherWriterStatus.
func (vss *IVssBackupComponents) FreeWriterStatus() error {
	code, _, _ := syscall.Syscall(vss.getVTable().freeWriterStatus, 1, uintptr(unsafe.Pointer(vss)), 0, 0)
	return CreateVSSError("IVssBackupComponents.FreeWriterStatus", code)
}

// The BackupComplete method causes VSS to generate a BackupComplete event, which signals writers that the backup process has completed.
func (vss *IVssBackupComponents) BackupComplete() (*IVssAsync, error) {
	var async *IVssAsync
	code, _, _ := syscall.Syscall(vss.getVTable().backupComplete, 2, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&async)), 0)
	if code != 0 {
		return nil, CreateVSSError("IVssBackupComponents.BackupComplete", code)
	}
	return async, nil
}

// The IsVolumeSupported method determines whether the specified provider supports shadow copies on the specified volume or remote file share.
// A zero providerID (GUID_NULL) checks whether any provider supports it.
func (vss *IVssBackupComponents) IsVolumeSupported(providerID ole.GUID, drive string) (bool, error) {
	var isSupported uint32
	var code uintptr

	volumeNamePointer, err := syscall.UTF16PtrFromString(drive)
	if err != nil {
		return false, err
	}

	if runtime.GOARCH == "386" {
		id := (*[4]uintptr)(unsafe.Pointer(&providerID))
		code, _, _ = syscall.Syscall9(vss.getVTable().isVolumeSupported, 7, uintptr(unsafe.Pointer(vss)), id[0], id[1], id[2], id[3], uintptr(unsafe.Pointer(volumeNamePointer)), uintptr(unsafe.Pointer(&isSupported)), 0, 0)
	} else if runtime.GOARCH == "arm64" {
		id := (*[2]uintptr)(unsafe.Pointer(&providerID))
		code, _, _ = syscall.Syscall6(vss.getVTable().isVolumeSupported, 5, uintptr(unsafe.Pointer(vss)), id[0], id[1], uintptr(unsafe.Pointer(volumeNamePointer)), uintptr(unsafe.Pointer(&isSupported)), 0)
	} else {
		code, _, _ = syscall.Syscall6(vss.getVTable().isVolumeSupported, 4, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&providerID)), uintptr(unsafe.Pointer(volumeNamePointer)), uintptr(unsafe.Pointer(&isSupported)), 0, 0)
	}
	return isSupported != 0, CreateVSSError("IVssBackupComponents.IsVolumeSupported", code)
}

// The StartSnapshotSet method creates a new, empty shadow copy set.
func (vss *IVssBackupComponents) StartSnapshotSet(snapshotID *ole.GUID) error {
	code, _, _ := syscall.Syscall(vss.getVTable().startSnapshotSet, 2, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(snapshotID)), 0)
	return CreateVSSError("IVssBackupComponents.StartSnapshotSet", code)
}

// The AddToSnapshotSet method adds an original volume or original remote file share to the shadow copy set.
// A zero providerID (GUID_NULL) lets VSS choose the provider.
func (vss *IVssBackupComponents) AddToSnapshotSet(drive string, providerID ole.GUID, snapshotID *ole.GUID) error {
	var code uintptr
	volumeName := syscall.StringToUTF16Ptr(drive)
	if runtime.GOARCH == "386" {
		id := (*[4]uintptr)(unsafe.Pointer(&providerID))
		code, _, _ = syscall.Syscall9(vss.getVTable().addToSnapshotSet, 7, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(volumeName)), id[0], id[1], id[2], id[3], uintptr(unsafe.Pointer(snapshotID)), 0, 0)
	} else if runtime.GOARCH == "arm64" {
		id := (*[2]uintptr)(unsafe.Pointer(&providerID))
		code, _, _ = syscall.Syscall6(vss.getVTable().addToSnapshotSet, 5, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(volumeName)), id[0], id[1], uintptr(unsafe.Pointer(snapshotID)), 0)
	} else {
		code, _, _ = syscall.Syscall6(vss.getVTable().addToSnapshotSet, 4, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(volumeName)), uintptr(unsafe.Pointer(&providerID)), uintptr(unsafe.Pointer(snapshotID)), 0, 0)
	}
	return CreateVSSError("IVssBackupComponents.AddToSnapshotSet", code)
}

// The SetContext method sets the context for subsequent shadow copy-related operations.
func (vss *IVssBackupComponents) SetContext(context VssContext) error {
	code, _, _ := syscall.Syscall(vss.getVTable().setContext, 2, uintptr(unsafe.Pointer(vss)), uintptr(context), 0)
	return CreateVSSError("IVssBackupComponents.SetContext", code)
}

// The SetBackupState method defines an overall configuration for a backup operation.
func (vss *IVssBackupComponents) SetBackupState(selectComponents bool, backupBootableState bool, state VssBackup, partialFileSupport bool) error {
	var bSelectComponents, bBackupBootableState, bPartialFileSupport uint32
	if selectComponents {
		bSelectComponents = 1
	}
	if backupBootableState {
		bBackupBootableState = 1
	}
	if partialFileSupport {
		bPartialFileSupport = 1
	}

	code, _, _ := syscall.Syscall6(vss.getVTable().setBackupState, 4, uintptr(unsafe.Pointer(vss)), uintptr(bSelectComponents), uintptr(bBackupBootableState), uintptr(state), uintptr(bPartialFileSupport), 0)
	return CreateVSSError("IVssBackupComponents.SetBackupState", code)
}

// The PrepareForBackup method will cause VSS to generate a PrepareForBackup event, signaling writers to prepare for an upcoming backup operation. This makes a requester's Backup Components Document available to writers.
func (vss *IVssBackupComponents) PrepareForBackup() (*IVssAsync, error) {
	var async *IVssAsync
	code, _, _ := syscall.Syscall(vss.getVTable().prepareForBackup, 2, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&async)), 0)
	if err := CreateVSSError("IVssBackupComponents.PrepareForBackup", code); err != nil {
		return nil, err
	}
	return async, nil
}

// Commits all shadow copies in this set simultaneously.
func (vss *IVssBackupComponents) DoSnapshotSet() (*IVssAsync, error) {
	var async *IVssAsync
	code, _, _ := syscall.Syscall(vss.getVTable().doSnapshotSet, 2, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&async)), 0)
	if err := CreateVSSError("IVssBackupComponents.DoSnapshotSet", code); err != nil {
		return nil, err
	}
	return async, nil
}

// The GetSnapshotProperties method gets the properties of the specified shadow copy.
func (vss *IVssBackupComponents) GetSnapshotProperties(snapshotSetID ole.GUID, properties *VssSnapshotProperties) error {
	var code uintptr
	if runtime.GOARCH == "386" {
		id := (*[4]uintptr)(unsafe.Pointer(&snapshotSetID))
		code, _, _ = syscall.Syscall6(vss.getVTable().getSnapshotProperties, 6, uintptr(unsafe.Pointer(vss)), id[0], id[1], id[2], id[3], uintptr(unsafe.Pointer(properties)))
	} else if runtime.GOARCH == "arm64" {
		id := (*[2]uintptr)(unsafe.Pointer(&snapshotSetID))
		code, _, _ = syscall.Syscall6(vss.getVTable().getSnapshotProperties, 4, uintptr(unsafe.Pointer(vss)), id[0], id[1], uintptr(unsafe.Pointer(properties)), 0, 0)
	} else {
		code, _, _ = syscall.Syscall(vss.getVTable().getSnapshotProperties, 3, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&snapshotSetID)), uintptr(unsafe.Pointer(properties)))
	}
	return CreateVSSError("IVssBackupComponents.GetSnapshotProperties", code)
}

// The Query method returns the completed shadow copies in the current context. Microsoft allows only GUID_NULL and VSS_OBJECT_NONE for the queried object.
func (vss *IVssBackupComponents) Query() (*IVssEnumObject, error) {
	var queriedObjectID ole.GUID
	var snapshots *IVssEnumObject
	var code uintptr
	if runtime.GOARCH == "386" {
		id := (*[4]uintptr)(unsafe.Pointer(&queriedObjectID))
		code, _, _ = syscall.Syscall9(vss.getVTable().query, 8, uintptr(unsafe.Pointer(vss)), id[0], id[1], id[2], id[3], uintptr(VSS_OBJECT_NONE), uintptr(VSS_OBJECT_SNAPSHOT), uintptr(unsafe.Pointer(&snapshots)), 0)
	} else if runtime.GOARCH == "arm64" {
		id := (*[2]uintptr)(unsafe.Pointer(&queriedObjectID))
		code, _, _ = syscall.Syscall6(vss.getVTable().query, 6, uintptr(unsafe.Pointer(vss)), id[0], id[1], uintptr(VSS_OBJECT_NONE), uintptr(VSS_OBJECT_SNAPSHOT), uintptr(unsafe.Pointer(&snapshots)))
	} else {
		code, _, _ = syscall.Syscall6(vss.getVTable().query, 5, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&queriedObjectID)), uintptr(VSS_OBJECT_NONE), uintptr(VSS_OBJECT_SNAPSHOT), uintptr(unsafe.Pointer(&snapshots)), 0)
	}
	return snapshots, CreateVSSError("IVssBackupComponents.Query", code)
}

// The DeleteSnapshots method deletes one or more shadow copies or a shadow copy set.
func (vss *IVssBackupComponents) DeleteSnapshots(snapshotID ole.GUID) (ole.GUID, bool, error) {
	VSS_OBJECT_SNAPSHOT := 3
	deleted := int32(0)

	var deletedGUID ole.GUID
	var code uintptr

	if runtime.GOARCH == "386" {
		id := (*[4]uintptr)(unsafe.Pointer(&snapshotID))
		code, _, _ = syscall.Syscall9(vss.getVTable().deleteSnapshots, 9, uintptr(unsafe.Pointer(vss)), id[0], id[1], id[2], id[3], uintptr(VSS_OBJECT_SNAPSHOT), uintptr(1), uintptr(unsafe.Pointer(&deleted)), uintptr(unsafe.Pointer(&deletedGUID)))
	} else if runtime.GOARCH == "arm64" {
		id := (*[2]uintptr)(unsafe.Pointer(&snapshotID))
		code, _, _ = syscall.Syscall9(vss.getVTable().deleteSnapshots, 7, uintptr(unsafe.Pointer(vss)), id[0], id[1], uintptr(VSS_OBJECT_SNAPSHOT), uintptr(1), uintptr(unsafe.Pointer(&deleted)), uintptr(unsafe.Pointer(&deletedGUID)), 0, 0)
	} else {
		code, _, _ = syscall.Syscall6(vss.getVTable().deleteSnapshots, 6, uintptr(unsafe.Pointer(vss)), uintptr(unsafe.Pointer(&snapshotID)), uintptr(VSS_OBJECT_SNAPSHOT), uintptr(1), uintptr(unsafe.Pointer(&deleted)), uintptr(unsafe.Pointer(&deletedGUID)))
	}
	return deletedGUID, code != 0, CreateVSSError("IVssBackupComponents.DeleteSnapshots", code)
}
