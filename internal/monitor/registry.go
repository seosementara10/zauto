package monitor

import (
	"log"
	"os/exec"
	"sync"
)

var (
	registryMu sync.Mutex
	owned      []*exec.Cmd
)

// RegisterScrcpy tracks a scrcpy process started by zauto so KillOwnedScrcpy
// can stop only our mirrors, not every scrcpy.exe on the system.
func RegisterScrcpy(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	registryMu.Lock()
	owned = append(owned, cmd)
	registryMu.Unlock()
	go reapWhenDone(cmd)
}

func reapWhenDone(cmd *exec.Cmd) {
	_ = cmd.Wait()
	registryMu.Lock()
	defer registryMu.Unlock()
	for i, c := range owned {
		if c == cmd {
			owned = append(owned[:i], owned[i+1:]...)
			break
		}
	}
}

// KillOwnedScrcpy stops only scrcpy processes that zauto started in this session.
func KillOwnedScrcpy() {
	registryMu.Lock()
	defer registryMu.Unlock()
	if len(owned) == 0 {
		return
	}
	n := 0
	for _, cmd := range owned {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
			n++
		}
	}
	owned = nil
	if n > 0 {
		log.Printf("Mirror: %d scrcpy zauto ditutup", n)
	}
}
