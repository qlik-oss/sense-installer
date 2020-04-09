package api

import "github.com/otiai10/copy"

//copy source directory to destination
func CopyDirectory(source string, dest string) error {
	return copy.Copy(source, dest)
}
