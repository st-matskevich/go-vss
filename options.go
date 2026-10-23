package vss

import (
	ole "github.com/go-ole/go-ole"
)

// SnapshotterOption configures optional behavior of Snapshotter.CreateSnapshot.
type SnapshotterOption func(*snapshotterOptions)

type snapshotterOptions struct {
	bootable             bool
	initCOMSecurity      bool
	systemProvider       bool
	withoutWriterClasses []ole.GUID
	inspectWriters       func([]WriterMetadata) error
}

func collectOptions(opts []SnapshotterOption) snapshotterOptions {
	var o snapshotterOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// WithBootable marks the created snapshot so that it can be exported as a
// bootable volume. It is passed through to IVssBackupComponents::SetBackupState.
func WithBootable() SnapshotterOption {
	return func(o *snapshotterOptions) { o.bootable = true }
}

// WithCOMSecurity initializes process-wide COM security via CoInitializeSecurity
// so that VSS writers running under restricted service accounts (such as
// NETWORK SERVICE or LOCAL SERVICE) are allowed to call back into this process
// through the IVssWriterCallback interface.
//
// Without it, those callbacks are denied with E_ACCESSDENIED and VSS logs
// event 8194 ("Unexpected error querying for the IVssWriterCallback interface.
// hr = 0x80070005, Access is denied."). Backups usually still succeed, so this
// is primarily about silencing that event-log noise.
//
// CoInitializeSecurity is process-wide and can be set only once per process and
// cannot be undone. Enable this only when go-vss owns the process, or when the
// host has not already configured COM security. If COM security was already
// configured (RPC_E_TOO_LATE), CreateSnapshot keeps the existing configuration
// instead of failing.
//
// See "Security Considerations for Requesters":
// https://learn.microsoft.com/en-us/windows/win32/vss/security-considerations-for-requestors
func WithCOMSecurity() SnapshotterOption {
	return func(o *snapshotterOptions) { o.initCOMSecurity = true }
}

// WithSystemProvider creates the snapshot with the system provider (Microsoft
// Software Shadow Copy provider 1.0) instead of letting VSS pick one.
//
// Left to pick, VSS prefers any hardware or software provider over the system
// provider. Inside Hyper-V guests that is the Hyper-V IC Software Shadow Copy
// Provider, and when it can't serve an in-guest backup AddToSnapshotSet fails
// with VSS_E_UNEXPECTED_PROVIDER_ERROR.
func WithSystemProvider() SnapshotterOption {
	return func(o *snapshotterOptions) { o.systemProvider = true }
}

// WithoutWriterClasses creates the snapshot without the writers of the given
// classes. It calls DisableWriterClasses before GatherWriterMetadata, so those
// writers receive no events for the snapshot. Their files are still in the
// snapshot.
func WithoutWriterClasses(classIDs ...ole.GUID) SnapshotterOption {
	return func(o *snapshotterOptions) { o.withoutWriterClasses = append(o.withoutWriterClasses, classIDs...) }
}

// WithWriterMetadata calls inspect with the metadata of the writers taking
// part, after VSS gathers it and before the snapshot set is created. When
// inspect returns an error, CreateSnapshot stops and returns it wrapped.
//
// No writer can be left out at that point, because DisableWriterClasses has no
// effect after GatherWriterMetadata. To leave a writer out, stop and call
// CreateSnapshot again with WithoutWriterClasses.
func WithWriterMetadata(inspect func(writers []WriterMetadata) error) SnapshotterOption {
	return func(o *snapshotterOptions) { o.inspectWriters = inspect }
}
