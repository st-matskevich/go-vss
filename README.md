# Golang VSS CGO-free module
Windows API bindings for the `Volume Shadow Copy Service` in Golang for 32-bit, 64-bit and ARM64 systems. Enables the user to duplicate entire drives during runtime without any file access issues. The API bindings are accompanied by a simple CLI tool that creates and symlinks Shadow Copies of a given drive. Uses syscalls and go-ole to avoid CGo.
## Build
You can either import the vss API bindings into your project or use the CLI application.

Cross-compilation is supported for 32-bit, 64-bit and ARM64 systems. The following commands will build the CLI application for the respective system:
```shell
make build_win32 # 32-bit x86
make build_win64 # 64-bit x86
make build_arm64 # 64-bit ARM
```

Regular build on Windows:
```shell
make build
```

## Usage
```sh
./vss -h
```
```
usage:  vss [options]
  -D string
        Drive letter to copy (example: C:\)
  -S string
        Path of symlink folder
  -bootable
        Created snapshot can be exported as a bootable volume
  -comsec
        Initialize COM security so VSS writers call back succeeds
  -timeout int
        Snapshot creation timeout in seconds (min 180) (default 180)
```

## Docs
Official MS Docs: https://docs.microsoft.com/en-us/windows/win32/api/vss/