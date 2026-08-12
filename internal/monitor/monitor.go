package monitor

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"zauto/internal/adb"
)

// Options layout for scrcpy windows on the PC monitor.
type Options struct {
	ProjectRoot string
	MaxSize     int
	StartX      int
	StartY      int
}

func DefaultOptions(projectRoot string) Options {
	return Options{
		ProjectRoot: projectRoot,
		MaxSize:     520,
		StartX:      20,
		StartY:      40,
	}
}

var (
	scrcpyCacheMu sync.RWMutex
	scrcpyCache   = map[string]string{}
)

// FindScrcpy locates scrcpy.exe in PATH or tools/scrcpy/.
func FindScrcpy(projectRoot string) (string, error) {
	scrcpyCacheMu.RLock()
	if p, ok := scrcpyCache[projectRoot]; ok {
		scrcpyCacheMu.RUnlock()
		return p, nil
	}
	scrcpyCacheMu.RUnlock()

	if p, err := exec.LookPath("scrcpy"); err == nil {
		scrcpyCacheMu.Lock()
		scrcpyCache[projectRoot] = p
		scrcpyCacheMu.Unlock()
		return p, nil
	}
	toolsDir := filepath.Join(projectRoot, "tools", "scrcpy")
	var found string
	_ = filepath.WalkDir(toolsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "scrcpy.exe") {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if found != "" {
		scrcpyCacheMu.Lock()
		scrcpyCache[projectRoot] = found
		scrcpyCacheMu.Unlock()
		return found, nil
	}
	return "", fmt.Errorf("scrcpy tidak ditemukan — unduh ke tools/scrcpy/ dari https://github.com/Genymobile/scrcpy/releases")
}

// ScrcpyDir returns the directory containing scrcpy.exe.
func ScrcpyDir(scrcpy string) string {
	return filepath.Dir(scrcpy)
}

// Start opens scrcpy windows for all serials. cleanExisting kills leftover scrcpy first (farm mode).
func Start(serials []string, opts Options, cleanExisting bool) ([]*exec.Cmd, error) {
	if len(serials) == 0 {
		return nil, fmt.Errorf("tidak ada serial untuk mirror")
	}
	if cleanExisting {
		KillOwnedScrcpy()
	}

	scrcpy, err := FindScrcpy(opts.ProjectRoot)
	if err != nil {
		return nil, err
	}
	scrcpyDir := filepath.Dir(scrcpy)
	log.Printf("Mirror: scrcpy — %d HP", len(serials))

	tiles := ComputeTiles(720, 1600, opts.MaxSize, len(serials), opts.StartX, opts.StartY)
	var (
		mu   sync.Mutex
		cmds []*exec.Cmd
		wg   sync.WaitGroup
	)
	for i, serial := range serials {
		wg.Add(1)
		go func(idx int, ser string) {
			defer wg.Done()
			time.Sleep(time.Duration(idx) * 250 * time.Millisecond)
			tile := tiles[idx]
			cmd, err := launchScrcpyAt(scrcpy, scrcpyDir, ser, idx+1, tile, opts)
			if err != nil {
				log.Printf("Mirror GAGAL %s: %v", ser, err)
				return
			}
			mu.Lock()
			cmds = append(cmds, cmd)
			mu.Unlock()
			log.Printf("Mirror [%d] %s -> (%d, %d) %dx%d", idx+1, ser, tile.X, tile.Y, tile.W, tile.H)
		}(i, serial)
	}
	wg.Wait()

	if len(cmds) == 0 {
		return nil, fmt.Errorf("tidak ada mirror yang berhasil distart")
	}
	return cmds, nil
}

// Stop closes all scrcpy processes started by this process.
func Stop(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
}

// Run starts mirror for all connected devices and waits until Enter.
func Run(opts Options) error {
	devices, err := adb.ListDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("tidak ada HP terhubung — cek USB dan adb devices")
	}

	fmt.Printf("=== Mirror %d HP ke monitor ===\n", len(devices))
	cmds, err := Start(devices, opts, false)
	if err != nil {
		return err
	}

	fmt.Println("\nMirror aktif. Jalankan automation di terminal lain: .\\zauto --no-mirror")
	fmt.Print("\nTekan Enter untuk tutup semua mirror...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')

	Stop(cmds)
	fmt.Println("\nSemua mirror ditutup.")
	return nil
}
