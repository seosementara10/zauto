//go:build windows

package monitor

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const processCommandLineInformation = 60

type unicodeString struct {
	Length        uint16
	MaximumLength uint16
	_             [4]byte
	Buffer        uintptr
}

var (
	modntdll                      = windows.NewLazySystemDLL("ntdll.dll")
	procNtQueryInformationProcess = modntdll.NewProc("NtQueryInformationProcess")
)

func processCommandLine(pid uint32) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)

	var info unicodeString
	var returnLength uint32
	r0, _, e1 := procNtQueryInformationProcess.Call(
		uintptr(h),
		uintptr(processCommandLineInformation),
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Sizeof(info)),
		uintptr(unsafe.Pointer(&returnLength)),
	)
	if r0 != 0 {
		if e1 != syscall.Errno(0) {
			return "", e1
		}
		return "", fmt.Errorf("NtQueryInformationProcess failed: 0x%x", r0)
	}
	if info.Buffer == 0 || info.Length == 0 {
		return "", fmt.Errorf("empty command line for pid %d", pid)
	}
	n := int(info.Length / 2)
	if n <= 0 {
		return "", fmt.Errorf("invalid command line length for pid %d", pid)
	}
	buf := make([]uint16, n)
	for i := 0; i < n; i++ {
		buf[i] = *(*uint16)(unsafe.Pointer(info.Buffer + uintptr(i*2)))
	}
	return syscall.UTF16ToString(buf), nil
}

func scrcpyProcessPIDs() []int {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snap)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	var pids []int
	for err := windows.Process32First(snap, &pe); err == nil; err = windows.Process32Next(snap, &pe) {
		if !strings.EqualFold(windows.UTF16ToString(pe.ExeFile[:]), "scrcpy.exe") {
			continue
		}
		pids = append(pids, int(pe.ProcessID))
	}
	return pids
}

// CountScrcpyProcesses returns the number of running scrcpy.exe instances.
func CountScrcpyProcesses() int {
	return len(scrcpyProcessPIDs())
}
