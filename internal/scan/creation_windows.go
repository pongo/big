//go:build windows

package scan

import (
	"io/fs"
	"syscall"
	"time"
)

func (osFS) CreationTime(_ string, info fs.FileInfo) (time.Time, bool) {
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return time.Time{}, false
	}
	return time.Unix(0, data.CreationTime.Nanoseconds()), true
}
