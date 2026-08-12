package monitor

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"zauto/internal/adb"
	"zauto/internal/uiutil"
)

const defaultDashboardPort = 8765

// Dashboard serves all device screens in one browser window.
type Dashboard struct {
	devices []string
	port    int
	server  *http.Server
	mu      sync.Mutex
}

// RunDashboard opens a unified grid view in the browser (blocking until Ctrl+C).
func RunDashboard(port int) error {
	devices, err := adb.ListDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("tidak ada HP terhubung")
	}
	d, err := newDashboard(devices, port)
	if err != nil {
		return err
	}
	url := d.url()
	log.Printf("Dashboard: %s (%d HP)", url, len(devices))
	uiutil.OpenAppWindow(url)
	fmt.Printf("\n=== Monitor unified — 1 jendela browser ===\n")
	fmt.Printf("Buka: %s\n", url)
	fmt.Println("Tekan Ctrl+C untuk stop")
	return d.server.ListenAndServe()
}

// StartDashboard runs the unified dashboard in background (for automation + mirror).
func StartDashboard(port int) (*Dashboard, error) {
	devices, err := adb.ListDevices()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("tidak ada HP terhubung")
	}
	d, err := newDashboard(devices, port)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := d.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Dashboard: %v", err)
		}
	}()
	time.Sleep(300 * time.Millisecond)
	uiutil.OpenAppWindow(d.url())
	log.Printf("Dashboard unified: %s (%d HP)", d.url(), len(devices))
	return d, nil
}

func (d *Dashboard) Stop() {
	if d.server != nil {
		_ = d.server.Close()
	}
}

func (d *Dashboard) url() string {
	return fmt.Sprintf("http://127.0.0.1:%d", d.port)
}

func newDashboard(devices []string, port int) (*Dashboard, error) {
	if port <= 0 {
		port = defaultDashboardPort
	}
	d := &Dashboard{devices: devices, port: port}
	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleIndex)
	for i, serial := range devices {
		idx := i
		ser := serial
		mux.HandleFunc(fmt.Sprintf("/shot/%d", idx), func(w http.ResponseWriter, r *http.Request) {
			d.handleShot(w, ser)
		})
	}
	d.server = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	return d, nil
}

func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML(d.devices))
}

func (d *Dashboard) handleShot(w http.ResponseWriter, serial string) {
	client := &adb.Client{Serial: serial, Timeout: 8 * time.Second}
	png, err := client.ScreenshotPNG()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}

func dashboardHTML(devices []string) string {
	n := len(devices)
	cells := ""
	for i, serial := range devices {
		short := serial
		if len(short) > 12 {
			short = serial[len(serial)-12:]
		}
		cells += fmt.Sprintf(`
		<div class="cell">
			<div class="label">HP%d <span class="serial">%s</span></div>
			<img id="img%d" src="/shot/%d" alt="HP%d">
		</div>`, i+1, short, i, i, i+1)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8">
<title>zauto monitor</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: #1a1a2e; color: #eee; font-family: Segoe UI, sans-serif; min-height: 100vh; }
  header { padding: 12px 20px; background: #16213e; border-bottom: 1px solid #0f3460; }
  header h1 { font-size: 18px; font-weight: 600; }
  header p { font-size: 12px; color: #888; margin-top: 4px; }
  .grid { display: flex; flex-wrap: wrap; gap: 12px; padding: 16px; justify-content: center; }
  .cell { background: #16213e; border-radius: 8px; padding: 8px; border: 1px solid #0f3460; flex: 1 1 280px; max-width: 380px; }
  .label { font-size: 13px; font-weight: 600; margin-bottom: 6px; text-align: center; }
  .serial { color: #7eb8da; font-weight: 400; font-size: 11px; }
  .cell img { width: 100%%; height: auto; border-radius: 4px; background: #000; display: block; min-height: 400px; object-fit: contain; }
</style>
</head>
<body>
<header>
  <h1>zauto — monitor unified</h1>
  <p>Semua HP dalam satu layar · refresh otomatis</p>
</header>
<div class="grid">%s</div>
<script>
const n = %d;
function refresh() {
  for (let i = 0; i < n; i++) {
    const img = document.getElementById('img' + i);
    if (img) img.src = '/shot/' + i + '?t=' + Date.now();
  }
}
setInterval(refresh, 800);
</script>
</body>
</html>`, cells, n)
}
