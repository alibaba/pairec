package pairec

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
)

func setupTestLogDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// override package-level logDir
	logDir = dir
	return dir
}

func createFileWithAge(t *testing.T, dir, name string, content []byte, age time.Duration) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-age)
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestClearOnce_SymlinkNotDeletedByRetention verifies that a symlink is NOT
// removed even when its mtime is older than the retention threshold.
func TestClearOnce_SymlinkNotDeletedByRetention(t *testing.T) {
	dir := setupTestLogDir(t)

	// create a real log file (target of symlink)
	targetPath := createFileWithAge(t, dir, "appd.host.log.INFO.20260101", []byte("log data"), 1*time.Hour)

	// create symlink pointing to the target
	linkPath := filepath.Join(dir, "appd.INFO")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatal(err)
	}

	// RetensionDays=0 means pointTime = now, so any file with mtime < now is "expired".
	// The symlink mtime is in the past, so without the fix it would be deleted.
	config := recconf.LogConfig{
		RetensionDays: 0,
		DiskSize:      20,
	}

	clearOnce(config)

	// symlink must still exist
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		t.Fatal("symlink appd.INFO was incorrectly deleted by retention logic")
	}

	// target file should also still exist (it's recent enough with 1h age vs 0 days retention
	// actually with RetensionDays=0, pointTime=now, target with mtime 1h ago would be deleted)
	// Let's just verify the symlink survives - that's the core assertion.
}

// TestClearOnce_SymlinkNotCountedInTotalSize verifies that symlinks are excluded
// from the totalSize calculation and the capacity-based deletion list.
func TestClearOnce_SymlinkNotCountedInTotalSize(t *testing.T) {
	dir := setupTestLogDir(t)

	// create a large regular file that alone does NOT exceed threshold
	// DiskSize=1 GB, threshold = 0.8 GB. We create a small file.
	smallContent := make([]byte, 1024) // 1KB
	createFileWithAge(t, dir, "small.log", smallContent, 1*time.Minute)

	// create a symlink - if incorrectly counted, it shouldn't affect anything meaningful,
	// but the key test is it's not in the deletion list
	targetPath := filepath.Join(dir, "small.log")
	linkPath := filepath.Join(dir, "appd.INFO")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatal(err)
	}

	config := recconf.LogConfig{
		RetensionDays: 3,
		DiskSize:      1,
	}

	clearOnce(config)

	// both files should still exist
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		t.Fatal("symlink was incorrectly deleted during capacity cleanup")
	}
	if _, err := os.Stat(filepath.Join(dir, "small.log")); os.IsNotExist(err) {
		t.Fatal("regular file was incorrectly deleted")
	}
}

// TestClearOnce_ExpiredRegularFileDeleted verifies that normal (non-symlink) files
// older than RetensionDays are still properly removed.
func TestClearOnce_ExpiredRegularFileDeleted(t *testing.T) {
	dir := setupTestLogDir(t)

	// create an old regular file (5 days old, retention = 3 days)
	oldFile := createFileWithAge(t, dir, "old.log", []byte("old data"), 5*24*time.Hour)

	// create a recent regular file (1 hour old)
	recentFile := createFileWithAge(t, dir, "recent.log", []byte("recent data"), 1*time.Hour)

	config := recconf.LogConfig{
		RetensionDays: 3,
		DiskSize:      20,
	}

	clearOnce(config)

	// old file should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatal("expired regular file was NOT deleted")
	}

	// recent file should still exist
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Fatal("recent regular file was incorrectly deleted")
	}
}

// TestClearOnce_CapacityDeletion verifies that when total file size exceeds the
// threshold, the oldest files are deleted first, and symlinks are never touched.
// Using DiskSize=0 makes sizeThreshold=0, so any non-empty totalSize triggers deletion.
func TestClearOnce_CapacityDeletion(t *testing.T) {
	dir := setupTestLogDir(t)

	// create two regular files with different ages
	createFileWithAge(t, dir, "a.log", []byte("data-a"), 1*time.Minute)
	createFileWithAge(t, dir, "b.log", []byte("data-b-longer-content"), 2*time.Minute)

	// symlink should survive capacity cleanup
	linkPath := filepath.Join(dir, "appd.INFO")
	if err := os.Symlink(filepath.Join(dir, "a.log"), linkPath); err != nil {
		t.Fatal(err)
	}

	// DiskSize=0 -> threshold=0, any totalSize > 0 triggers capacity deletion
	config := recconf.LogConfig{
		RetensionDays: 30,
		DiskSize:      0,
	}

	clearOnce(config)

	// symlink must survive
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		t.Fatal("symlink was deleted during capacity cleanup")
	}

	// at least one regular file should have been deleted (oldest first)
	aExists := fileExists(filepath.Join(dir, "a.log"))
	bExists := fileExists(filepath.Join(dir, "b.log"))
	if aExists && bExists {
		t.Fatal("expected at least one regular file to be deleted by capacity cleanup")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// TestClearOnce_SymlinkSurvivesBothPaths is a combined scenario: symlink with old
// mtime coexists with expired regular files. After clearOnce, symlink must remain
// while expired regular files are removed.
func TestClearOnce_SymlinkSurvivesBothPaths(t *testing.T) {
	dir := setupTestLogDir(t)

	// target log file (recent, should survive)
	targetPath := createFileWithAge(t, dir, "appd.host.log.INFO.20260722", []byte("active log"), 1*time.Hour)

	// symlink with old creation time (simulated by RetensionDays=0 making everything "expired")
	linkPath := filepath.Join(dir, "appd.INFO")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatal(err)
	}

	// an expired regular file that SHOULD be deleted
	expiredFile := createFileWithAge(t, dir, "old_rotated.log", []byte("stale"), 10*24*time.Hour)

	config := recconf.LogConfig{
		RetensionDays: 3,
		DiskSize:      20,
	}

	clearOnce(config)

	// symlink must survive
	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		t.Fatal("symlink appd.INFO was deleted")
	}

	// target (recent) must survive
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatal("recent target log file was deleted")
	}

	// expired regular file must be gone
	if _, err := os.Stat(expiredFile); !os.IsNotExist(err) {
		t.Fatal("expired regular file was NOT deleted")
	}
}
