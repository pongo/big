//go:build !windows

package scan

import (
	"io/fs"
	"time"
)

func (osFS) CreationTime(_ string, _ fs.FileInfo) (time.Time, bool) {
	return time.Time{}, false
}
