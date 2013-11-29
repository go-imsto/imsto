package storage

import (
	"os"
	"syscall"
)

// FLock is a file-based lock
type FLock struct {
	fh *os.File
}

// NewFLock creates new Flock-based lock (unlocked first)
func NewFLock(path string) (FLock, error) {
	fh, err := os.Create(path)
	if err != nil {
		return FLock{}, err
	}
	return FLock{fh: fh}, nil
}

// Lock acquires the lock, blocking
func (lock FLock) Lock() error {
	return syscall.Flock(int(lock.fh.Fd()), syscall.LOCK_EX)
}

// TryLock acquires the lock, non-blocking
func (lock FLock) TryLock() (bool, error) {
	err := syscall.Flock(int(lock.fh.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	switch err {
	case nil:
		return true, nil
	case syscall.EWOULDBLOCK:
		return false, nil
	}
	return false, err
}

// Unlock releases the lock
func (lock FLock) Unlock() error {
	lock.fh.Close()
	return syscall.Flock(int(lock.fh.Fd()), syscall.LOCK_UN)
}
