package main

import (
	"errors"
	"log"
	"syscall"
	"unsafe"
)

type (
	BOOL          uint32
	BOOLEAN       byte
	BYTE          byte
	DWORD         uint32
	DWORD64       uint64
	HANDLE        uintptr
	HLOCAL        uintptr
	LARGE_INTEGER int64
	LONG          int32
	LPVOID        uintptr
	SIZE_T        uintptr
	UINT          uint32
	ULONG_PTR     uintptr
	ULONGLONG     uint64
	WORD          uint16
)

// 常量
const (
	PROCESS_ALL_ACCESS = 0x001F0FFF
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
)

var (
	FindWindowW               = user32.NewProc("FindWindowW")
	GetWindowThreadProcessIdW = user32.NewProc("GetWindowThreadProcessId")
	OpenProcessW              = kernel32.NewProc("OpenProcess")
	CLoseHandleW              = kernel32.NewProc("CloseHandle")
	ReadProcessMemoryW        = kernel32.NewProc("ReadProcessMemory")
	GetExitCodeProcessW       = kernel32.NewProc("GetExitCodeProcess")
	WriteProcessMemoryW       = kernel32.NewProc("WriteProcessMemory")
	VirtualAllocExW           = kernel32.NewProc("VirtualAllocEx")
	VirtualFreeExW            = kernel32.NewProc("VirtualFreeEx")
	CreateRemoteThreadW       = kernel32.NewProc("CreateRemoteThread")
	WaitForSingleObjectW      = kernel32.NewProc("WaitForSingleObject")
)

func FindWindow(className, windowName string) HANDLE {
	classNameptr, _ := syscall.UTF16PtrFromString(className)
	windowNameptr, _ := syscall.UTF16PtrFromString(windowName)
	r1, _, err := FindWindowW.Call(
		uintptr(unsafe.Pointer(classNameptr)),
		uintptr(unsafe.Pointer(windowNameptr)),
	)
	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return HANDLE(r1)
}

func GetWindowThreadProcessId(hWnd HANDLE, lpdwProcessId *DWORD) DWORD {
	r1, _, err := GetWindowThreadProcessIdW.Call(
		uintptr(hWnd),
		uintptr(unsafe.Pointer(lpdwProcessId)),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return DWORD(r1)
}

func OpenProcess(dwDesiredAccess DWORD, bInheritHandle BOOL, dwProcessId DWORD) HANDLE {
	r1, _, err := OpenProcessW.Call(
		uintptr(dwDesiredAccess),
		uintptr(bInheritHandle),
		uintptr(dwProcessId),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return HANDLE(r1)
}

func CloseHandle(hObject HANDLE) BOOL {
	r1, _, err := CLoseHandleW.Call(
		uintptr(hObject),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return BOOL(r1)
}

func ReadProcessMemory(hProcess HANDLE, lpBaseAddress LPVOID, lpBuffer *LPVOID, nSize SIZE_T, lpNumberOfBytesRead *SIZE_T) BOOL {
	r1, _, err := ReadProcessMemoryW.Call(
		uintptr(hProcess),
		uintptr(lpBaseAddress),
		uintptr(unsafe.Pointer(lpBuffer)),
		uintptr(nSize),
		uintptr(unsafe.Pointer(lpNumberOfBytesRead)),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return BOOL(r1)
}

func GetExitCodeProcess(hProcess HANDLE, lpExitCode *DWORD) BOOL {
	r1, _, err := GetExitCodeProcessW.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(lpExitCode)),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return BOOL(r1)
}

func WriteProcessMemory(hProcess HANDLE, lpBaseAddress LPVOID, lpBuffer LPVOID, nSize SIZE_T, lpNumberOfBytesWritten *SIZE_T) BOOL {
	r1, _, err := WriteProcessMemoryW.Call(
		uintptr(hProcess),
		uintptr(lpBaseAddress),
		uintptr(lpBuffer),
		uintptr(nSize),
		uintptr(unsafe.Pointer(lpNumberOfBytesWritten)),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return BOOL(r1)
}

func VirtualAllocEx(hProcess HANDLE, lpAddress LPVOID, dwSize SIZE_T, flAllocationType DWORD, flProtect DWORD) LPVOID {
	r1, _, err := VirtualAllocExW.Call(
		uintptr(hProcess),
		uintptr(lpAddress),
		uintptr(dwSize),
		uintptr(flAllocationType),
		uintptr(flProtect),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return LPVOID(r1)
}

func VituralFreeEx(hProcess HANDLE, lpAddress LPVOID, dwSize SIZE_T, dwFreeType DWORD) BOOL {
	r1, _, err := VirtualFreeExW.Call(
		uintptr(hProcess),
		uintptr(lpAddress),
		uintptr(dwSize),
		uintptr(dwFreeType),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return BOOL(r1)
}

func CreateRemoteThread(hProcess HANDLE, lpThreadAttributes LPVOID, dwStackSize SIZE_T, lpStartAddress LPVOID, lpParameter LPVOID, dwCreationFlags DWORD, lpThreadId *DWORD) HANDLE {
	r1, _, err := CreateRemoteThreadW.Call(
		uintptr(hProcess),
		uintptr(lpThreadAttributes),
		uintptr(dwStackSize),
		uintptr(lpStartAddress),
		uintptr(lpParameter),
		uintptr(dwCreationFlags),
		uintptr(unsafe.Pointer(lpThreadId)),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return HANDLE(r1)
}

func WaitForSingleObject(hHandle HANDLE, dwMilliseconds DWORD) DWORD {
	r1, _, err := WaitForSingleObjectW.Call(
		uintptr(hHandle),
		uintptr(dwMilliseconds),
	)

	if !errors.Is(err, syscall.Errno(0)) {
		log.Panic(err)
	}

	return DWORD(r1)
}
