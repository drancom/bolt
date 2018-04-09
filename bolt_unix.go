// +build !windows,!plan9,!solaris

package bolt

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// flock acquires an advisory lock on a file descriptor.
func flock(db *DB, mode os.FileMode, exclusive bool, timeout time.Duration) error {
	var t time.Time
	for {
		// If we're beyond our timeout then return an error.
		// This can only occur after we've attempted a flock once.
		if t.IsZero() {
			t = time.Now()
		} else if timeout > 0 && time.Since(t) > timeout {
			return ErrTimeout
		}
		flag := syscall.LOCK_SH
		if exclusive {
			flag = syscall.LOCK_EX
		}

		// Otherwise attempt to obtain an exclusive lock.
		err := syscall.Flock(int(db.file.Fd()), flag|syscall.LOCK_NB)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK {
			return err
		}

		// Wait for a bit and try again.
		time.Sleep(50 * time.Millisecond)
	}
}

// funlock releases an advisory lock on a file descriptor.
func funlock(db *DB) error {
	return syscall.Flock(int(db.file.Fd()), syscall.LOCK_UN)
}

func mmapRW(db *DB, sz int) error {
	err := mmapOpen(db, sz, syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED| syscall.MAP_HASSEMAPHORE |db.MmapFlags)
	return err
}

func mmap(db *DB, sz int) error {
	err := mmapOpen(db, sz, syscall.PROT_READ,                      syscall.MAP_SHARED|db.MmapFlags)
	return err
}

// mmap memory maps a DB's data file.
func mmapOpen(db *DB, sz int, prot int, flag int) error {
	// Map the data file to memory.
	b, err := syscall.Mmap(int(db.file.Fd()), 0, sz, prot, flag)
	if err != nil {
		return err
	}

	// Advise the kernel that the mmap is accessed randomly.
	if err := madvise(b, syscall.MADV_RANDOM); err != nil {
		return fmt.Errorf("madvise: %s", err)
	}

	// Save the original byte slice and convert to a byte array pointer.
	db.dataref = b
	db.data = (*[maxMapSize]byte)(unsafe.Pointer(&b[0]))
	db.datasz = sz
	return nil
}

func writeMmap(db *DB, src []byte, off int64) (n int, err error) {
	sz := len(src)
	dest := db.data

	if sz > db.datasz {
		sz = db.datasz
		n = 0
		return n, fmt.Errorf("sz is bigger than datasz")
	}

	copy(dest[off:off+int64(sz)], src)

	if err != nil {
		return 0, err
	}
	n = sz

	return n,nil
}

// func msync (b []byte, len uintptr) error {
//	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, b, len, syscall.MS_SYNC)
//	if errno != 0 {
	//	return syscall.Errno(errno)
//	}
//	return nil
//}

// munmap unmaps a DB's data file from memory.
func munmap(db *DB) error {
	// Ignore the unmap if we have no mapped data.
	if db.dataref == nil {
		return nil
	}

	// Unmap using the original byte slice.
	err := syscall.Munmap(db.dataref)
	db.dataref = nil
	db.data = nil
	db.datasz = 0
	return err
}

// NOTE: This function is copied from stdlib because it is not available on darwin.
func madvise(b []byte, advice int) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_MADVISE, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(advice))
	if e1 != 0 {
		err = e1
	}
	return
}
