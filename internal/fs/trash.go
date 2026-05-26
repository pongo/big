package fs

import "github.com/hymkor/trash-go"

func TrashPath(path string) error {
	return trash.Throw(path)
}
