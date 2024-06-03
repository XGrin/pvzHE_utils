package main

import (
	"log"
	"unsafe"
)

// 寄存器
const (
	// 通用寄存器
	EAX = 0
	ECX = 1
	EDX = 2
	EBX = 3
	ESP = 4
	EBP = 5
	ESI = 6
	EDI = 7
)

const (
	MEM_COMMIT             = 0x00001000
	PAGE_EXECUTE_READWRITE = 0x40
)

type Code struct {
	page      uint16
	code      []byte
	length    uint16
	calls_pos []uint16
}

func (c *Code) asm_init() {
	c.length = 0
	// 清空
	c.calls_pos = c.calls_pos[:0]
}

func asm_add_byte(c *Code, value byte) {
	c.code[c.length] = value
	c.length++
}

func asm_add[T interface{}](c *Code, value T) {
	// 转换为字节数组
	temp := ToBytes(value)
	// 添加到code
	for _, v := range temp {
		asm_add_byte(c, v)
	}
}

func asm_push_byte(c *Code, value byte) {
	asm_add_byte(c, 0x68)
	asm_add_byte(c, value)
}

func asm_push[T interface{}](c *Code, value T) {
	asm_add_byte(c, 0x68)
	asm_add(c, value)
}

func asm_mov_exx[T interface{}](c *Code, reg uint8, value T) {
	asm_add_byte(c, 0xB8+reg)
	asm_add(c, value)
}

func asm_mov_exx_dword_ptr(c *Code, reg uint8, value uint32) {
	asm_add_byte(c, 0x8b)
	asm_add_byte(c, 0x05+reg*8)
	asm_add[uint32](c, value)

}

func asm_mov_exx_dword_ptr_exx_add(c *Code, reg uint8, value uint32) {
	asm_add_byte(c, 0x8B)
	asm_add_byte(c, 0x80+reg*(8+1))
	if reg == ESP {
		asm_add_byte(c, 0x24)
	}
	asm_add[uint32](c, value)

}

func asm_push_exx(c *Code, reg uint8) {
	asm_add_byte(c, 0x50+reg)
}

func asm_pop_exx(c *Code, reg uint8) {
	asm_add_byte(c, 0x58+reg)
}

func asm_mov_exx_exx(c *Code, reg1 uint8, reg2 uint8) {
	asm_add_byte(c, 0x8b)
	asm_add_byte(c, 0xC0+reg1*8+reg2)
}

func asm_call(c *Code, addr uint32) {
	asm_add_byte(c, 0xE8)
	c.calls_pos = append(c.calls_pos, c.length)
	asm_add[uint32](c, addr)

}

func asm_ret(c *Code) {
	asm_add_byte(c, 0xC3)
}

func asm_code_inject(c *Code, handle HANDLE) {
	addr := VirtualAllocEx(handle, LPVOID(0), SIZE_T(c.length), MEM_COMMIT, PAGE_EXECUTE_READWRITE)
	for i := 0; i < len(c.calls_pos); i++ {
		pos := c.calls_pos[i]
		tmp := bytesTo[int32](c.code[pos : pos+4])
		tmp = tmp - (int32(addr) + int32(pos) + 4)
		// 将tmp转换回字节切片
		temp := ToBytes(tmp)
		// 将temp写入code
		for j := 0; j < 4; j++ {
			c.code[int(pos)+j] = temp[j]
		}
	}
	var write_size SIZE_T = 0
	ret := WriteProcessMemory(handle, addr, LPVOID(unsafe.Pointer(&c.code[0])), SIZE_T(c.length), &write_size)
	if ret == 0 || write_size != SIZE_T(c.length) {
		log.Println("写入失败")
		VituralFreeEx(handle, addr, 0, 0x8000)
		return
	}
	var temp DWORD = 0
	thread := CreateRemoteThread(handle, LPVOID(0), 0, addr, LPVOID(0), 0, &temp)
	if thread == 0 {
		log.Println("创建线程失败")
		VituralFreeEx(handle, addr, 0, 0x8000)
		return
	}
	WaitForSingleObject(thread, 0xFFFFFFFF)
	CloseHandle(thread)
	VituralFreeEx(handle, addr, 0, 0x8000)

}

func bytesTo[T interface{}](b []byte) T {
	return *(*T)(unsafe.Pointer(&b[0]))
}

// @title: ToBytes
// @description: 将任意类型转换为字节切片
// @param: val T 任意类型
// @return: []byte
func ToBytes[T interface{}](val T) []byte {
	size := unsafe.Sizeof(val)
	bytes := make([]byte, size)
	for i := 0; i < int(size); i++ {
		bytes[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i)))
	}
	return bytes
}
