package lock

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestFileLock_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lock := NewFileLock(testFile)

	// Test initial state
	if lock.IsLocked() {
		t.Error("New lock should not be locked")
	}

	// Test lock acquisition
	if err := lock.Lock(); err != nil {
		t.Errorf("Failed to acquire lock: %v", err)
	}

	if !lock.IsLocked() {
		t.Error("Lock should be locked after Lock()")
	}

	// Test unlock
	if err := lock.Unlock(); err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}

	if lock.IsLocked() {
		t.Error("Lock should not be locked after Unlock()")
	}
}

func TestFileLock_TryLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lock := NewFileLock(testFile)

	// Test try lock success
	if err := lock.TryLock(); err != nil {
		t.Errorf("TryLock should succeed: %v", err)
	}

	if !lock.IsLocked() {
		t.Error("Lock should be locked after TryLock()")
	}

	// Test unlock
	if err := lock.Unlock(); err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}
}

func TestFileLock_Concurrent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lock1 := NewFileLock(testFile)
	lock2 := NewFileLock(testFile)

	// Acquire first lock
	if err := lock1.Lock(); err != nil {
		t.Fatalf("Failed to acquire first lock: %v", err)
	}
	defer func() { _ = lock1.Unlock() }()

	// Try to acquire second lock (should fail)
	if err := lock2.TryLock(); err == nil {
		t.Error("Second TryLock should fail when file is already locked")
		_ = lock2.Unlock()
	}
}

func TestFileLock_LockWithTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lock1 := NewFileLock(testFile)
	lock2 := NewFileLock(testFile)

	// Acquire first lock
	if err := lock1.Lock(); err != nil {
		t.Fatalf("Failed to acquire first lock: %v", err)
	}

	// Try to acquire second lock with short timeout (should timeout)
	start := time.Now()
	err = lock2.LockWithTimeout(200 * time.Millisecond)
	duration := time.Since(start)

	if err == nil {
		t.Error("Second lock should timeout")
		_ = lock2.Unlock()
	}

	if duration < 150*time.Millisecond {
		t.Error("Lock should wait for timeout duration")
	}

	_ = lock1.Unlock()
}

func TestFileLock_DoubleOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lock := NewFileLock(testFile)

	// Test double lock
	if err := lock.Lock(); err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if err := lock.Lock(); err == nil {
		t.Error("Second Lock() should fail")
	}

	// Test unlock
	if err := lock.Unlock(); err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}

	// Test double unlock
	if err := lock.Unlock(); err == nil {
		t.Error("Second Unlock() should fail")
	}
}

func TestWithLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executed := false
	err = WithLock(testFile, 1*time.Second, func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("WithLock failed: %v", err)
	}

	if !executed {
		t.Error("Function should have been executed")
	}
}

func TestWithLock_FunctionError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = WithLock(testFile, 1*time.Second, func() error {
		return os.ErrInvalid
	})

	if err != os.ErrInvalid {
		t.Errorf("WithLock should return function error, got: %v", err)
	}
}

func TestWithQuickLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executed := false
	err = WithQuickLock(testFile, func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("WithQuickLock failed: %v", err)
	}

	if !executed {
		t.Error("Function should have been executed")
	}
}

func TestFileLock_ConcurrentWithLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var executed1, executed2 bool
	var wg sync.WaitGroup
	var startTime = time.Now()

	wg.Add(2)

	// First goroutine holds lock for 200ms
	go func() {
		defer wg.Done()
		err := WithLock(testFile, 2*time.Second, func() error {
			executed1 = true
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Errorf("First WithLock failed: %v", err)
		}
	}()

	// Second goroutine waits for lock
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond) // Ensure first goroutine gets lock first
		err := WithLock(testFile, 2*time.Second, func() error {
			executed2 = true
			return nil
		})
		if err != nil {
			t.Errorf("Second WithLock failed: %v", err)
		}
	}()

	wg.Wait()
	totalTime := time.Since(startTime)

	if !executed1 || !executed2 {
		t.Error("Both functions should have been executed")
	}

	// Second function should wait for first to complete
	if totalTime < 200*time.Millisecond {
		t.Error("Second function should wait for first to complete")
	}
}

func TestFileLock_LockFileCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "flock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	lockFile := testFile + ".lock"

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	lock := NewFileLock(testFile)

	// Lock file should not exist initially
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock file should not exist before locking")
	}

	// Acquire lock
	if err := lock.Lock(); err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Lock file should exist while locked
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file should exist while locked")
	}

	// Release lock
	if err := lock.Unlock(); err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}

	// Lock file should be cleaned up after unlock
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock file should be cleaned up after unlock")
	}
}

func TestNewFileLock(t *testing.T) {
	testPath := "/tmp/test.txt"
	lock := NewFileLock(testPath)

	if lock == nil {
		t.Error("NewFileLock should not return nil")
		return
	}

	if lock.path != testPath {
		t.Errorf("NewFileLock path = %v, want %v", lock.path, testPath)
	}

	if lock.acquired {
		t.Error("New lock should not be acquired")
	}

	if lock.file != nil {
		t.Error("New lock should not have file handle")
	}
}
