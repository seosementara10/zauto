package panel

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"zauto/internal/projectroot"
)

// Serve runs the panel HTTP server without a desktop window.
func Serve(opts Options) error {
	if opts.Port <= 0 {
		opts.Port = DefaultPort
	}

	prep, err := Prepare(context.Background(), opts)
	if err != nil {
		return err
	}
	srv := prep.Server
	srv.refreshDevices()
	srv.broadcastState()

	url := fmt.Sprintf("http://127.0.0.1:%d", opts.Port)
	go func() {
		log.Printf("Panel HTTP: %s", url)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Panel server: %v", err)
		}
	}()

	time.Sleep(300 * time.Millisecond)

	fmt.Printf("\n=== zauto Panel (browser) ===\n%s\n", url)
	if srv.panelDevEnabled() {
		fmt.Println("DEV aktif — edit internal/panel/web → auto reload (~1–2 detik)")
		fmt.Println("F12 DevTools untuk error JS / Network")
	}
	fmt.Println("Tekan Ctrl+C untuk stop")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Stop panel...")
	prep.Close()
	return nil
}

// ServeBrowserDev enables dev mode and runs headless panel for browser testing.
func ServeBrowserDev(args []string) error {
	fs := flag.NewFlagSet("dev-browser", flag.ExitOnError)
	configPath := fs.String("config", "config/config.json", "Path to config JSON")
	port := fs.Int("port", DefaultPort, "Panel HTTP port")
	openBrowser := fs.Bool("open", true, "Open default browser")
	if err := fs.Parse(args); err != nil {
		return err
	}

	root := projectroot.Find()
	if err := EnablePanelDevFlag(root); err != nil {
		return err
	}
	_ = os.Setenv("ZAUTO_PANEL_DEV", "1")

	if PanelInstanceActive() {
		fmt.Println("Menutup zautopanel.exe — browser mode pakai HTTP port, bukan jendela Wails.")
		stopDesktopPanel()
		time.Sleep(600 * time.Millisecond)
	}

	cfg := projectroot.ResolveConfig(*configPath)
	opts := Options{ProjectRoot: root, ConfigPath: cfg, Port: *port}
	if opts.Port <= 0 {
		opts.Port = DefaultPort
	}

	prep, err := Prepare(context.Background(), opts)
	if err != nil {
		return err
	}
	srv := prep.Server
	srv.refreshDevices()
	srv.broadcastState()

	url := fmt.Sprintf("http://127.0.0.1:%d", opts.Port)
	go func() {
		log.Printf("Panel HTTP (dev): %s", url)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Panel server: %v", err)
		}
	}()

	time.Sleep(400 * time.Millisecond)

	if *openBrowser {
		openBrowserURL(url)
	}

	fmt.Printf("\n=== zauto Panel DEV (browser) ===\n%s\n", url)
	fmt.Println("F12 = DevTools · Tab Log = server log · Ctrl+C = stop")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Stop panel...")
	prep.Close()
	return nil
}

func openBrowserURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("panel: buka browser gagal: %v — buka manual: %s", err, url)
	}
}

// RunServeCLI is the `zauto serve` entrypoint (headless HTTP only).
func RunServeCLI(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "config/config.json", "Path to config JSON")
	port := fs.Int("port", DefaultPort, "Panel HTTP port")
	dev := fs.Bool("dev", false, "Serve UI from disk with hot reload")
	openBrowser := fs.Bool("open", false, "Open default browser")
	if err := fs.Parse(args); err != nil {
		return err
	}

	root := projectroot.Find()
	if *dev {
		_ = EnablePanelDevFlag(root)
		_ = os.Setenv("ZAUTO_PANEL_DEV", "1")
	}

	cfg := projectroot.ResolveConfig(*configPath)
	opts := Options{ProjectRoot: root, ConfigPath: cfg, Port: *port}

	if *openBrowser {
		// Serve blocks until Ctrl+C; open browser after bind.
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowserURL(fmt.Sprintf("http://127.0.0.1:%d", opts.Port))
		}()
	}

	return Serve(opts)
}
