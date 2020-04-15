package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

func main() {
	srcPattern := flag.String("src-pattern", "", "Source file pattern")
	src := flag.String("src", "", "Source file or directory")
	dst := flag.String("dst", "", "Destination file or directory")
	flag.Parse()

	if *srcPattern != "" {
		if dstInfo, err := os.Lstat(*dst); err != nil {
			panic(err)
		} else if !dstInfo.IsDir() {
			panic(fmt.Errorf("%v must be a directory", *dst))
		}

		if matches, err := filepath.Glob(*srcPattern); err != nil {
			panic(err)
		} else {
			for _, match := range matches {
				srcInfo, err := os.Lstat(match)
				if err != nil {
					panic(err)
				}
				if srcInfo.IsDir() {
					if err := copy.Copy(match, *dst, copy.Options{
						OnSymlink: func(p string) copy.SymlinkAction {
							return copy.Skip
						},
					}); err != nil {
						panic(err)
					}
				} else if srcInfo.Mode().IsRegular() {
					if err := fcopy(match, filepath.Join(*dst, filepath.Base(match)), srcInfo); err != nil {
						panic(err)
					}
				}
			}
		}
	} else if *src != "" {
		srcInfo, err := os.Lstat(*src)
		if err != nil {
			panic(err)
		}
		if srcInfo.IsDir() {
			if err := copy.Copy(*src, *dst, copy.Options{
				OnSymlink: func(p string) copy.SymlinkAction {
					return copy.Skip
				},
			}); err != nil {
				panic(err)
			}
		} else if srcInfo.Mode().IsRegular() {
			finalDestination := *dst
			if dstInfo, err := os.Lstat(*dst); err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
			} else if dstInfo.IsDir() {
				finalDestination = filepath.Join(*dst, filepath.Base(*src))
			} else if dstInfo.Mode().IsRegular() {
				fmt.Println("WARNING: over-writing existing file: ", *dst)
				if err := os.Remove(*dst); err != nil {
					panic(err)
				}
			} else {
				panic(fmt.Errorf("not sure how to copy to this dst: %v", *dst))
			}
			if err := fcopy(*src, finalDestination, srcInfo); err != nil {
				panic(err)
			}
		}
	}
}

func fcopy(src, dest string, info os.FileInfo) (err error) {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}
