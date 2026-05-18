//go:build windows

package main

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

type cpuPlatformState struct {
	mu          sync.Mutex
	prevIdle    uint64
	prevKernel  uint64
	prevUser    uint64
	firstCall   bool
}

var (
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemTimes = kernel32.NewProc("GetSystemTimes")
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	ExtendedAvailVirtual uint64
}

func getSystemRAM() (totalMB, usedPercent float64) {
	procGlobalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")
	var ms memoryStatusEx
	ms.Length = uint32(unsafe.Sizeof(ms))
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	if ret == 0 {
		return 0, 0
	}
	totalMB = float64(ms.TotalPhys) / 1024 / 1024
	usedPercent = float64(ms.MemoryLoad)
	return
}

func (a *App) getCPUPercent() float64 {
	a.cpuPlatState.mu.Lock()
	defer a.cpuPlatState.mu.Unlock()

	var idleFT, kernelFT, userFT windows.Filetime
	ret, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleFT)),
		uintptr(unsafe.Pointer(&kernelFT)),
		uintptr(unsafe.Pointer(&userFT)),
	)
	if ret == 0 {
		return 0
	}

	idle := uint64(idleFT.LowDateTime) | (uint64(idleFT.HighDateTime) << 32)
	kernel := uint64(kernelFT.LowDateTime) | (uint64(kernelFT.HighDateTime) << 32)
	user := uint64(userFT.LowDateTime) | (uint64(userFT.HighDateTime) << 32)

	if a.cpuPlatState.firstCall {
		a.cpuPlatState.prevIdle = idle
		a.cpuPlatState.prevKernel = kernel
		a.cpuPlatState.prevUser = user
		a.cpuPlatState.firstCall = false
		return 0
	}

	idleDelta := idle - a.cpuPlatState.prevIdle
	kernelDelta := kernel - a.cpuPlatState.prevKernel
	userDelta := user - a.cpuPlatState.prevUser

	totalDelta := kernelDelta + userDelta

	a.cpuPlatState.prevIdle = idle
	a.cpuPlatState.prevKernel = kernel
	a.cpuPlatState.prevUser = user

	if totalDelta == 0 {
		return 0
	}

	return float64(totalDelta-idleDelta) / float64(totalDelta) * 100.0
}
