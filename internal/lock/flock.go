// Package lock provides file locking utilities to prevent concurrent access.
// It uses system-level file locking (flock) to ensure safe file operations.
package lock

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// FileLock represents a file lock that prevents concurrent access to a file.
// It creates a separate .lock file and uses flock system calls for locking.
type FileLock struct {
	file     *os.File // Lock file handle
	path     string   // Path to the file being locked
	acquired bool     // Whether the lock is currently held
}

// NewFileLock creates a new FileLock for the specified file path.
// The actual lock file will be created with a .lock extension.
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: path,
	}
}

// Lock acquires the file lock with a default timeout of 30 seconds.
// It's a convenience method that calls LockWithTimeout.
func (fl *FileLock) Lock() error {
	return fl.LockWithTimeout(30 * time.Second)
}

// LockWithTimeout attempts to acquire the file lock within the specified timeout.
// It will retry periodically until the lock is acquired or timeout is reached.
func (fl *FileLock) LockWithTimeout(timeout time.Duration) error {
	if fl.acquired {
		return fmt.Errorf("lock already acquired")
	}

	file, err := os.OpenFile(fl.path+".lock", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			fl.file = file
			fl.acquired = true

			pid := fmt.Sprintf("%d\n", os.Getpid())
			file.WriteString(pid)
			file.Sync()

			return nil
		}

		if err != syscall.EAGAIN && err != syscall.EACCES {
			file.Close()
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if time.Now().After(deadline) {
			file.Close()
			return fmt.Errorf("timeout waiting for lock")
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// TryLock attempts to acquire the lock immediately without waiting.
// Returns an error if the lock cannot be acquired right away.
func (fl *FileLock) TryLock() error {
	if fl.acquired {
		return fmt.Errorf("lock already acquired")
	}

	file, err := os.OpenFile(fl.path+".lock", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		if err == syscall.EAGAIN || err == syscall.EACCES {
			return fmt.Errorf("lock is already held by another process")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	fl.file = file
	fl.acquired = true

	pid := fmt.Sprintf("%d\n", os.Getpid())
	file.WriteString(pid)
	file.Sync()

	return nil
}

// Unlock releases the file lock and cleans up the lock file.
// The lock file is removed after the lock is released.
func (fl *FileLock) Unlock() error {
	if !fl.acquired {
		return fmt.Errorf("lock not acquired")
	}

	err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	fl.file.Close()
	os.Remove(fl.path + ".lock")

	fl.file = nil
	fl.acquired = false

	return nil
}

// IsLocked returns true if the lock is currently held by this instance.
func (fl *FileLock) IsLocked() bool {
	return fl.acquired
}

// WithLock is a convenience function that acquires a lock, executes a function,
// and automatically releases the lock when done.
func WithLock(path string, timeout time.Duration, fn func() error) error {
	lock := NewFileLock(path)

	if err := lock.LockWithTimeout(timeout); err != nil {
		return err
	}
	defer lock.Unlock()

	return fn()
}

// WithQuickLock is a convenience function for short operations with a 5-second timeout.
// It's equivalent to WithLock with a 5-second timeout.
func WithQuickLock(path string, fn func() error) error {
	return WithLock(path, 5*time.Second, fn)
}
