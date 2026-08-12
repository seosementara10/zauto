package panel

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	go func() {
		log.Printf("Panel HTTP: http://127.0.0.1:%d", opts.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Panel server: %v", err)
		}
	}()

	fmt.Printf("\n=== zauto Panel (headless) ===\nhttp://127.0.0.1:%d\n", opts.Port)
	fmt.Println("Tekan Ctrl+C untuk stop")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Stop panel...")
	prep.Close()
	return nil
}

// RunServeCLI is the `zauto serve` entrypoint (headless HTTP only).
func RunServeCLI(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "config/config.json", "Path to config JSON")
	port := fs.Int("port", DefaultPort, "Panel HTTP port")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := projectroot.Find()
	cfg := projectroot.ResolveConfig(*configPath)
	return Serve(Options{
		ProjectRoot: root,
		ConfigPath:  cfg,
		Port:        *port,
	})
}
