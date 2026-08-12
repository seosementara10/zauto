# Multi-Device Android Automation (Go)

Otomatisasi banyak HP Android via **ADB** + **Go** — dirancang untuk **100+ device** paralel dengan goroutine.

## Persyaratan

- **Go 1.22+** ([download](https://go.dev/dl/))
- [Android Platform Tools (ADB)](https://developer.android.com/tools/releases/platform-tools) di PATH
- HP Android dengan USB debugging aktif

## Build

```powershell
go build -o zauto.exe ./cmd/zauto
```

## Setup cepat

```powershell
adb devices
go build -o zauto.exe ./cmd/zauto
.\zauto --check
.\zauto --list-devices
.\zauto --dry-run
.\zauto --panel
```

> Di PowerShell wajib pakai `.\` di depan — `zauto.exe` saja tidak dikenali.

## Panel zauto (satu perintah)

```powershell
.\zauto --panel
```

Task **facebook_login** menjalankan flow lengkap:
1. Buka Facebook (`com.facebook.katana`)
2. Login otomatis (akun dari `data/accounts.txt`)
3. Logout via menu
4. Force-stop / tutup aplikasi

Satu jendela panel:
1. **Sidebar** — aktifkan HP (On/Off), centang skrip
2. **Preview** — layar HP yang aktif saja
3. **Jalankan** — start automation

`--farm` sama dengan `--panel`.

Scale **10+ HP** → STF:

```powershell
.\zauto --farm-stf
```

## Monitor alternatif (tanpa panel)

```powershell
.\zauto --monitor-windows   # scrcpy terpisah
.\zauto --monitor           # dashboard browser
```

## Scale ke 100 HP

1. Hubungkan semua HP via USB hub **powered**
2. Pastikan semua muncul di `adb devices`
3. Tambah akun di `data/accounts.txt` (1 baris = 1 HP)
4. Jalankan:

```powershell
.\zauto.exe --max-devices 100 --workers 100
```

## Arsitektur (4 lapisan)

```
Go (zauto)          → controller, scheduler, device pool, worker, retry, logging
ADB                 → install APK, start/stop app, screenshot, input, device mgmt
Appium              → automation UI Android (HTTP WebDriver)
UiAutomator2        → backend Appium: find / click / fill elemen UI
```

| Lapisan | Paket | Fungsi |
|---------|-------|--------|
| **Go** | `controller/`, `scheduler/`, `pool/`, `worker/`, `logging/` | Orkestrasi 100+ HP, concurrency, retry, log per device |
| **ADB** | `adb/` | `install`, `force-stop`, `am start`, `screencap`, `input tap/text`, `uiautomator dump` |
| **Driver** | `driver/` | Interface `UI` — pilih backend via config `automation.driver` |
| **Appium** | `appium/` | Client W3C ke server Appium 2 |
| **Engine** | `engine/` | Handlers task (tap, fill, login, register) |

Default driver: **`adb`** (tanpa server Appium). Untuk UI kompleks, set `"automation": { "driver": "appium" }` dan jalankan Appium server per device (port 4723, 4724, …).

## Struktur

```
cmd/zauto/           → entry point CLI
internal/
  controller/        → orchestrator (device filter, run, summary)
  scheduler/         → bounded concurrency across pool
  pool/              → device pool (idle/busy/error)
  worker/            → per-device loop + retry
  logging/           → per-device log files
  adb/               → ADB client (install, tap, screenshot, …)
  appium/            → Appium 2 W3C client
  driver/            → UI interface (adb dump | appium uiautomator2)
  config/            → JSON config + flows
  ui/                → XML hierarchy parser
  engine/            → session, executor, handlers
  data/              → names, emails loaders (legacy txt)
  store/             → PostgreSQL accounts, devices, assignments
config/
  config.json        → satu-satunya config produksi
data/                → names.txt, email.txt (accounts via PostgreSQL)
```

## Config (`config/config.json`)

Hanya **satu file config** untuk produksi. Tidak ada config test terpisah.

| Bagian | Fungsi |
|--------|--------|
| `max_devices` / `parallel_workers` | Scale hingga 100 HP (default 100) |
| `database.url` | PostgreSQL zauto (port **5433**, terpisah dari WY CMS) |
| `database.max_accounts_per_device` | Maks slot akun per HP (default 50) |
| `tasks` | Task yang dijalankan (saat ini: `facebook_login` → flow `facebook_login_logout`) |
| `automation.driver` | `adb` (default) atau `appium` |

**Driver UI** (`automation.driver`):

| Value | Keterangan |
|-------|------------|
| `adb` | Default — uiautomator dump + parse XML lokal, tanpa Appium |
| `appium` | Appium server + UiAutomator2 (stabil untuk find/click/fill) |

### Memulai sistem

Build sekali (butuh CGO + WebView2):

```powershell
.\build.ps1
```

Lalu:

```powershell
.\up.ps1              # Docker DB + panel
.\zautopanel.exe      # panel native langsung
.\zauto.exe open      # sama, lewat launcher
```

`zauto.exe` = CLI automation. `zautopanel.exe` = aplikasi desktop Wails (harus di-build dengan tag `desktop,production`).

### Database akun (PostgreSQL)

PostgreSQL port **5433** (terpisah WY CMS 5432). Schema & import awal otomatis saat panel pertama kali jalan.

Kelola akun & assign HP **di panel** (langkah 1 — Akun Facebook):

- **Import data/accounts.txt** — muat akun ke database
- **Assign HP ↔ Akun** — HP ke-1 → akun 1, HP ke-2 → akun 2 (colok USB dulu)

Format import legacy (`data/accounts.txt`):

```text
# name|password|email|profile_id|fanpage1|fanpage2|fanpage3
Nurhayati Guguk|Qwewenzqwewenz795172*||61592118521974|615931763399|61593132631889|61592753657118
```

- Login pakai **email** jika diisi, otherwise **profile_id** (Profile utama)
- **Assign HP ↔ Akun** di panel (HP ke-1 → akun 1, dst.)
- Panel/run otomatis ambil akun dari DB; HP tanpa assignment ditandai "belum assign"

```powershell
# 2 HP dengan assignment di DB
.\zauto.exe --max-devices 2 --workers 2

# 100 HP (butuh 100 baris di accounts.txt)
.\zauto.exe --max-devices 100 --workers 100
```

## CLI

```powershell
.\zauto --help
.\zauto --list-devices
.\zauto --monitor
.\zauto --dry-run
.\zauto --max-devices 2 --workers 2
```

## Tips stabilitas (100 HP)

- USB hub **powered** (bukan bus-powered saja)
- Aktifkan **Stay awake** di Developer options
- Label fisik HP = serial ADB
- Realme/Oppo: terima **Allow USB debugging** di setiap HP
- Cek log di `logs/`
