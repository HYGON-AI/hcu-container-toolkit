/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package configure

import (
	"io"
	"os"
	"path/filepath"
)

func backupFileIfExists(src, dst string) error {
	st, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if st.IsDir() {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, st.Mode())
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func restoreBackupIfExists(backup, dst string) error {
	if _, err := os.Stat(backup); err != nil {
		return err
	}
	return backupFileIfExists(backup, dst)
}
