package scan

import "fmt"

func FormatSize(size int64) string {
	const unit = int64(1024)

	switch {
	case size < unit:
		return fmt.Sprintf("%d B", size)
	case size < unit*unit:
		return fmt.Sprintf("%d KB", size/unit)
	case size < unit*unit*unit:
		return fmt.Sprintf("%d MB", size/(unit*unit))
	default:
		return fmt.Sprintf("%d GB", size/(unit*unit*unit))
	}
}
