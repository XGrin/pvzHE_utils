package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 定义一些常量和类型
const (
	maxTitleLength = 255
)

var (
	procEnumWindows   = user32.NewProc("EnumWindows")
	procGetWindowText = user32.NewProc("GetWindowTextW")
)

// @title: pvzWindow
// @description: pvz窗口结构体
type pvzWindow struct {
	// 窗口句柄
	Handle HANDLE
	// 进程ID
	Pid DWORD
	// 进程句柄
	ProcessHandle HANDLE
	// 内存锁
	memoryLock chan struct{}
	// 标题
	title string
}

type HWND uintptr

// 定义 EnumWindows 回调函数类型
type EnumWindowsProc func(HWND, uintptr) uintptr

// EnumWindows 函数
func EnumWindows(enumFunc EnumWindowsProc, lParam uintptr) error {
	ret, _, err := procEnumWindows.Call(
		syscall.NewCallback(enumFunc),
		lParam,
	)
	if ret == 0 {
		return err
	}
	return nil
}

// GetWindowText 函数
func GetWindowText(hwnd HWND) (string, error) {
	buf := make([]uint16, maxTitleLength)
	ret, _, err := procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		return "", err
	}
	return syscall.UTF16ToString(buf), nil
}

// 全局变量，用于存储是否找到匹配的窗口
var found bool
var searchString string

// 回调函数，用于处理每个枚举到的窗口
func enumWindowsCallback(hwnd HWND, lParam uintptr) uintptr {
	title, err := GetWindowText(hwnd)
	if err == nil && title != "" {
		if containsIgnoreCase(title, searchString) {
			found = true
			pvz.title = title
			return 0 // 停止枚举
		}
	}
	return 1 // 继续枚举
}

// containsIgnoreCase 函数，用于不区分大小写地判断子字符串是否存在
func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// CheckWindowTitle 函数，用于检查是否有包含指定字符串标题的窗口
func CheckWindowTitle(substr string) bool {
	found = false
	searchString = substr
	err := EnumWindows(enumWindowsCallback, 0)
	if err != nil {
		if !strings.Contains(err.Error(), "The operation completed successfully.") {
			log.Println("EnumWindows failed:", err)
		}
	}
	return found
}

func IsAdmin() (bool, error) {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false, err
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false, err
	}
	return member, nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

//使用io.Copy
func CopyFile(src, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	//获取源文件的权限
	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	//desFile, err := os.Create(des)  //无法复制源文件的所有权限
	desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm) //复制源文件的所有权限
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}

//使用ioutil.WriteFile()和ioutil.ReadFile()
func CopyFile2(src, des string) (written int64, err error) {
	//获取源文件的权限
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	fi, _ := srcFile.Stat()
	perm := fi.Mode()
	srcFile.Close()

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return 0, err
	}

	err = ioutil.WriteFile(des, input, perm)
	if err != nil {
		return 0, err
	}

	return int64(len(input)), nil
}

//使用os.Read()和os.Write()
func CopyFile3(src, des string, bufSize int) (written int64, err error) {
	if bufSize <= 0 {
		bufSize = 1 * 1024 * 1024 //1M
	}
	buf := make([]byte, bufSize)

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	//获取源文件的权限
	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	desFile, err := os.OpenFile(des, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	count := 0
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}

		if n == 0 {
			break
		}

		if wn, err := desFile.Write(buf[:n]); err != nil {
			return 0, err
		} else {
			count += wn
		}
	}

	return int64(count), nil
}

func CopyDir(srcPath, desPath string) error {
	//检查目录是否正确
	if srcInfo, err := os.Stat(srcPath); err != nil {
		return err
	} else {
		if !srcInfo.IsDir() {
			return errors.New("源路径不是一个正确的目录！")
		}
	}

	if desInfo, err := os.Stat(desPath); err != nil {
		return err
	} else {
		if !desInfo.IsDir() {
			return errors.New("目标路径不是一个正确的目录！")
		}
	}

	if strings.TrimSpace(srcPath) == strings.TrimSpace(desPath) {
		return errors.New("源路径与目标路径不能相同！")
	}

	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		//复制目录是将源目录中的子目录复制到目标路径中，不包含源目录本身
		if path == srcPath {
			return nil
		}

		//生成新路径
		destNewPath := strings.Replace(path, srcPath, desPath, -1)

		if !f.IsDir() {
			CopyFile(path, destNewPath)
		} else {
			if !FileIsExisted(destNewPath) {
				return MakeDir(destNewPath)
			}
		}

		return nil
	})

	return err
}

func IsDir(name string) bool {
	if info, err := os.Stat(name); err == nil {
		return info.IsDir()
	}
	return false
}

func FileIsExisted(filename string) bool {
	existed := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		existed = false
	}
	return existed
}

func MakeDir(dir string) error {
	if !FileIsExisted(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil { //os.ModePerm
			log.Println("MakeDir failed:", err)
			return err
		}
	}
	return nil
}

func (pvz *pvzWindow) CallSave() {
	cd := &Code{
		page:      256,
		code:      make([]byte, 1024),
		length:    0,
		calls_pos: make([]uint16, 0),
	}

	asm_mov_exx_dword_ptr(cd, ECX, 0x006A9EC0)
	asm_mov_exx_dword_ptr_exx_add(cd, ECX, 0x768)
	asm_push_exx(cd, ECX)
	asm_call(cd, 0x408C30)
	asm_ret(cd)
	asm_code_inject(cd, pvz.ProcessHandle)
}

// @title: pvzWindow::ReadMemory
// @description: 读取内存
// @param: readSize int 读取字节数
// @param: address ...int 内存地址(可以多级偏移)
// @return: interface{}
func (pvz *pvzWindow) ReadMemory(readSize int, address ...int) interface{} {
	if !pvz.IsValid() {
		log.Panic("窗口无效!")
	}

	// 加锁
	pvz.memoryLock <- struct{}{}
	defer func() {
		// 解锁
		<-pvz.memoryLock
	}()

	level := len(address)       // 偏移级数
	var offset LPVOID = 0       // 内存地址
	var buffer = new(LPVOID)    // 缓冲区
	var bytesRead = new(SIZE_T) // 读取字节数

	for i := 0; i < level; i++ {
		offset = *buffer + LPVOID(address[i])
		if i != level-1 {
			size := 4
			success := ReadProcessMemory(pvz.ProcessHandle, LPVOID(offset), buffer, SIZE_T(size), bytesRead)
			if success == 0 && *bytesRead != SIZE_T(size) {
				log.Panic("读取内存失败!")
			}
		} else {
			sucess := ReadProcessMemory(pvz.ProcessHandle, LPVOID(offset), buffer, SIZE_T(readSize), bytesRead)
			if sucess == 0 && *bytesRead != SIZE_T(readSize) {
				log.Panic("读取内存失败!")
			}
		}
	}

	// log.Printf("读取内存, 地址 %v, 字节数 %d, 结果 %d.", address, readSize, *buffer)
	return *buffer
}

// @title: pvzWindow::WriteMemory
// @description: 写入内存
// @param: writeBuffer []byte 要写入的字节
// @param: writeSzie int 写入字节数
// @param: address ...int 内存地址(可以多级偏移)
// @return: void
func (pvz *pvzWindow) WriteMemory(writeBuffer []byte, writeSzie int, address ...int) {
	if !pvz.IsValid() {
		log.Panic("窗口无效!")
	}

	// 加锁
	pvz.memoryLock <- struct{}{}
	defer func() {
		// 解锁
		<-pvz.memoryLock
	}()

	level := len(address)       // 偏移级数
	var offset LPVOID = 0       // 内存地址
	var buffer = new(LPVOID)    // 缓冲区
	var bytesRead = new(SIZE_T) // 读取字节数

	for i := 0; i < level; i++ {
		offset = *buffer + LPVOID(address[i])
		if i != level-1 {
			size := 4
			success := ReadProcessMemory(pvz.ProcessHandle, LPVOID(offset), buffer, SIZE_T(size), bytesRead)
			if success == 0 && *bytesRead != SIZE_T(size) {
				log.Panic("读取内存失败!")
			}
		} else {
			bytesWrite := new(SIZE_T)
			sucess := WriteProcessMemory(pvz.ProcessHandle, LPVOID(offset), LPVOID(unsafe.Pointer(&writeBuffer[0])), SIZE_T(writeSzie), bytesWrite)
			if sucess == 0 && *bytesWrite != SIZE_T(writeSzie) {
				log.Panic("写入内存失败!")
			}
		}
	}

	// log.Printf("写入内存, 地址 %v, 字节数 %d, 结果 %v.", address, writeSzie, writeBuffer)
}

// @title: pvzWindow::isValid
// @description: 判断窗口是否有效
// @return: bool
func (pvz *pvzWindow) IsValid() bool {
	if pvz.Handle == 0 {
		return false
	}
	exit_code := new(DWORD)
	GetExitCodeProcess(pvz.ProcessHandle, exit_code)
	if *exit_code == 259 {
		return true
	} else {
		return false
	}
}

// @title: pvzWindow::GetGameUI
// @description: 获取游戏界面类型
// @return: int 1: 主界面, 2: 选卡, 3: 正常游戏/战斗, 4: 僵尸进屋, 7: 模式选择, -1: 不可用
func (pvz *pvzWindow) GetGameUI() int {
	if !pvz.IsValid() {
		return -1
	}
	return int(pvz.ReadMemory(4, 0x6a9ec0, 0x7FC).(LPVOID))
}

func (pvz *pvzWindow) PlayMusic(id int) {
	if !pvz.IsValid() {
		log.Panic("窗口无效!")
	}
	cd := &Code{
		page:      256,
		code:      make([]byte, 1024),
		length:    0,
		calls_pos: make([]uint16, 0),
	}
	asm_mov_exx(cd, EDI, id)
	asm_mov_exx_dword_ptr(cd, EAX, 0x6A9EC0)
	asm_mov_exx_dword_ptr_exx_add(cd, EAX, 0x83c)
	asm_call(cd, 0x0045b750)
	asm_ret(cd)
	asm_code_inject(cd, pvz.ProcessHandle)

}

func (pvz *pvzWindow) GetMusicID() int {
	if !pvz.IsValid() {
		log.Panic("窗口无效!")
	}
	return int(pvz.ReadMemory(4, 0x6a9ec0, 0x83c, 0x8).(LPVOID))
}
