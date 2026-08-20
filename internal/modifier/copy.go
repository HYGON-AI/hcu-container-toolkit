/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"bytes"
	"errors"
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"hcu-container-toolkit/internal/oci"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
)

type copyModifier struct {
	logger logger.Interface
}

// Gives a number indicating the device driver to be used to access the passed device
func major(device uint64) uint64 {
	return (device >> 8) & 0xfff
}

// Gives a number that serves as a flag to the device driver for the passed device
func minor(device uint64) uint64 {
	return (device & 0xff) | ((device >> 12) & 0xfff00)
}

func mkdev(major int64, minor int64) uint32 {
	return uint32(((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff))
}

func CopyFile(source string, dest string) error {
	si, err := os.Lstat(source)
	if err != nil {
		return err
	}

	st, ok := si.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("could not convert to syscall.Stat_t")
	}

	uid := int(st.Uid)
	gid := int(st.Gid)
	modeType := si.Mode() & os.ModeType

	// Handle symlinks
	if modeType == os.ModeSymlink {
		target, err := os.Readlink(source)
		if err != nil {
			return err
		}
		if _, err := os.Lstat(dest); err == nil {
			if err := os.Remove(dest); err != nil {
				return fmt.Errorf("failed to remove existing file: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check if destination exists: %w", err)
		}
		if err := os.Symlink(target, dest); err != nil {
			return err
		}
	}

	// Handle device files
	if modeType == os.ModeDevice {
		devMajor := int64(major(uint64(st.Rdev)))
		devMinor := int64(minor(uint64(st.Rdev)))
		mode := uint32(si.Mode() & os.ModePerm)
		if si.Mode()&os.ModeCharDevice != 0 {
			mode |= syscall.S_IFCHR
		} else {
			mode |= syscall.S_IFBLK
		}
		if err := syscall.Mknod(dest, mode, int(mkdev(devMajor, devMinor))); err != nil {
			return err
		}
	}

	// Handle regular files
	if si.Mode().IsRegular() {
		err = copyInternal(source, dest)
		if err != nil {
			return err
		}
	}

	// Chown the file
	if err := os.Lchown(dest, uid, gid); err != nil {
		return err
	}

	// Chmod the file
	if !(modeType == os.ModeSymlink) {
		if err := os.Chmod(dest, si.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func copyInternal(source, dest string) (retErr error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		err := df.Close()
		if retErr == nil {
			retErr = err
		}
	}()

	_, err = io.Copy(df, sf)
	return err
}
func RemoveSymlinkOrDirectory(source string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return fmt.Errorf("failed to lstat source: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		err := os.Remove(source)
		if err != nil {
			return fmt.Errorf("failed to remove symlink: %w", err)
		}
	} else if info.IsDir() {
		err := os.RemoveAll(source)
		if err != nil {
			return fmt.Errorf("failed to remove directory: %w", err)
		}
	} else {
		return fmt.Errorf("source is neither a symlink nor a directory: %s", source)
	}
	return nil
}

func CopyDirectory(srcDir, dstDir string) error {
	RemoveSymlinkOrDirectory(dstDir)
	fi, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("could not convert to syscall.Stat_t")
	}
	if err := os.MkdirAll(dstDir, fi.Mode()); err != nil {
		return err
	}

	if err := os.Lchown(dstDir, int(st.Uid), int(st.Gid)); err != nil {
		return err
	}
	if err := os.Chmod(dstDir, fi.Mode()); err != nil {
		return err
	}
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)
		if info.IsDir() {
			st, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return fmt.Errorf("could not convert to syscall.Stat_t")
			}
			uid := int(st.Uid)
			gid := int(st.Gid)
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return err
			}
			if err := os.Lchown(dstPath, uid, gid); err != nil {
				return err
			}
			if err := os.Chmod(dstPath, info.Mode()); err != nil {
				return err
			}
			return nil
		}
		return CopyFile(srcPath, dstPath)
	})
}

func NewCopyModifier(logger logger.Interface) (oci.SpecModifier, error) {
	m := copyModifier{
		logger: logger,
	}
	return &m, nil
}

type VFS interface {
	Lstat(name string) (os.FileInfo, error)
	Readlink(name string) (string, error)
}

type osVFS struct{}

func (o osVFS) Lstat(name string) (os.FileInfo, error) { return os.Lstat(name) }
func (o osVFS) Readlink(name string) (string, error)   { return os.Readlink(name) }

// IsNotExist tells you if err is an error that implies that either the path
// accessed does not exist (or path components don't exist). This is
// effectively a more broad version of os.IsNotExist.
func IsNotExist(err error) bool {
	// Check that it's not actually an ENOTDIR, which in some cases is a more
	// convoluted case of ENOENT (usually involving weird paths).
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) || errors.Is(err, syscall.ENOENT)
}
func SecureJoinVFS(root, unsafePath string, vfs VFS) (string, error) {
	// Use the os.* VFS implementation if none was specified.
	if vfs == nil {
		vfs = osVFS{}
	}

	unsafePath = filepath.FromSlash(unsafePath)
	var path bytes.Buffer
	n := 0
	for unsafePath != "" {
		if n > 255 {
			return "", &os.PathError{Op: "SecureJoin", Path: root + string(filepath.Separator) + unsafePath, Err: syscall.ELOOP}
		}

		if v := filepath.VolumeName(unsafePath); v != "" {
			unsafePath = unsafePath[len(v):]
		}

		// Next path component, p.
		i := strings.IndexRune(unsafePath, filepath.Separator)
		var p string
		if i == -1 {
			p, unsafePath = unsafePath, ""
		} else {
			p, unsafePath = unsafePath[:i], unsafePath[i+1:]
		}

		// Create a cleaned path, using the lexical semantics of /../a, to
		// create a "scoped" path component which can safely be joined to fullP
		// for evaluation. At this point, path.String() doesn't contain any
		// symlink components.
		cleanP := filepath.Clean(string(filepath.Separator) + path.String() + p)
		if cleanP == string(filepath.Separator) {
			path.Reset()
			continue
		}
		fullP := filepath.Clean(root + cleanP)

		// Figure out whether the path is a symlink.
		_, err := vfs.Lstat(fullP)
		if err != nil && !IsNotExist(err) {
			return "", err
		}
		// Treat non-existent path components the same as non-symlinks (we
		// can't do any better here).
		path.WriteString(p)
		path.WriteRune(filepath.Separator)
	}

	// We have to clean path.String() here because it may contain '..'
	// components that are entirely lexical, but would be misleading otherwise.
	// And finally do a final clean to ensure that root is also lexically
	// clean.
	fullP := filepath.Clean(string(filepath.Separator) + path.String())
	return filepath.Clean(root + fullP), nil
}

func (m copyModifier) Modify(spec *specs.Spec) error {
	locator := lookup.NewDirectoryLocator(
		lookup.WithLogger(m.logger),
		lookup.WithCount(1),
		lookup.WithSearchPaths("/usr/local", "/opt"),
	)
	candidate := "hyhal"
	located, err := locator.Locate(candidate)
	if err != nil {
		m.logger.Warningf("Could not locate %v: %v", candidate, err)
		return nil
	}
	if len(located) == 0 {
		m.logger.Warningf("Missing %v", candidate)
		return nil
	}
	m.logger.Debugf("Located %v as %v", candidate, located)
	for _, path := range located {
		container_path, err := SecureJoinVFS(spec.Root.Path, path, nil)
		if err != nil {
			return err
		}
		return CopyDirectory(path, container_path)
	}

	return nil
}
