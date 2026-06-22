package vss

// SnapshotterOption configures optional behavior of Snapshotter.CreateSnapshot.
type SnapshotterOption func(*snapshotterOptions)

type snapshotterOptions struct {
	bootable        bool
	initCOMSecurity bool
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
