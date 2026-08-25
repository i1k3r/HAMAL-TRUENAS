<div align="center">

<img src="brand/hamal-logo-dark.svg" alt="HAMAL Logo" width="320" />

# HAMAL

**Point-to-point local file transfer. The carrier for your digital cargo.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](Dockerfile)
[![TrueNAS SCALE](https://img.shields.io/badge/TrueNAS%20SCALE-24.10%2B%20(Electric%20Eel)-0078D7?style=flat-square)](docs/truenas-installation-guide.md)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

</div>

---

## What is HAMAL?

**HAMAL** (Turkish for *porter* / *carrier*) is a fast, lightweight, self-hosted web application designed for frictionless, temporary file transfers across your local network (LAN / Wi-Fi).

In traditional culture, a *hamal* carries goods and heavy loads from one place to another. In **HAMAL**, the cargo is digital: files move directly between devices on your local network without intermediate cloud storage, permanent user accounts, advertising, or telemetry.

```
+---------------+                    +----------------+                    +-----------------+
|               |  1. Scans QR Code  |                |  2. Streams File   |                 |
|  Room Creator | -----------------> |  HAMAL Server  | <----------------- |   Participant   |
|   (Desktop)   |                    | (Local Network)|                    | (Mobile/Laptop) |
+---------------+                    +----------------+                    +-----------------+
```

---

## Screenshots

<div align="center">

### Home / Landing Page
*Start temporary transfer rooms with customizable TTL and optional PIN security.*
<br/>
<img src="docs/screenshots/home.jpg" alt="HAMAL Home UI" width="900" />

<br/><br/>

### Creator Dispatch Dashboard
*Real-time QR code with instant expansion lightbox, parcel dropzone, countdown timer, and secure file manifests.*
<br/>
<img src="docs/screenshots/creator.jpg" alt="HAMAL Creator UI" width="900" />

</div>

---

## Key Features

- 🚀 **Zero Setup for Participants**: Scan a QR code or open a local link to immediately upload or download files.
- ⏱️ **Auto-Expiring Rooms**: Rooms automatically expire and clean up files after a configurable TTL (5 minutes to 24 hours).
- 🔒 **PIN Protection**: Optional 4–8 digit PIN with exponential backoff and lockout to prevent brute-force attacks.
- 📦 **True Streaming I/O**: Multi-gigabyte transfers stream directly to disk without exhausting server RAM.
- 🎨 **Warm Courier Aesthetics**: Clean, modern interface in dark and light modes with warm amber courier accents.
- 🔍 **Interactive QR Lightbox**: One-click smooth zoom for scanning QR codes from across the room.
- 🛡️ **Self-Hosted & Private**: Zero cloud dependencies, zero external analytics, zero tracking.
- 🔒 **Hardened Container Security**: Non-root UID/GID `10001`, read-only root filesystem, in-memory `tmpfs` `/tmp`, and `no-new-privileges`.

---

## TrueNAS SCALE Installation

HAMAL runs natively on **TrueNAS SCALE 24.10 ("Electric Eel") or later** using the built-in **Install via YAML** feature.

📖 **Complete Guide:** [**TrueNAS SCALE Step-by-Step Installation Guide**](docs/truenas-installation-guide.md)

### Quick Deployment (Install via YAML)

1. **Create Dataset**: Create a dataset on your TrueNAS pool (e.g. `/mnt/tank/appdata/hamal`).
2. **Set Permissions**: In TrueNAS WebUI **Datasets** -> **Edit Permissions** / **Edit ACL**, grant **Read + Write** permissions to UID `10001` and GID `10001`.
3. **Install App**:
   - Go to **Apps** -> **Discover Apps** -> **⋮ (top-right menu)** -> **Install via YAML**.
   - Set **Application Name** to `hamal`.
   - Paste the YAML from [`truenas/compose.yaml`](truenas/compose.yaml):

```yaml
services:
  hamal:
    image: ghcr.io/i1k3r/hamal-truenas:latest
    restart: unless-stopped
    read_only: true
    user: "10001:10001"
    security_opt:
      - no-new-privileges:true
    ports:
      - "7700:7700"
    volumes:
      # Set this to the absolute path of your TrueNAS ZFS dataset
      - /mnt/tank/appdata/hamal:/data
    tmpfs:
      - /tmp:size=64M,mode=1777
    environment:
      - TZ=Europe/Istanbul
```

4. Click **Install**. Once running, access the WebUI at `http://<TRUENAS_IP>:7700`.

---

## Quick Start (Generic Docker Compose)

1. Clone the repository:
   ```bash
   git clone https://github.com/i1k3r/HAMAL-TRUENAS.git
   cd HAMAL-TRUENAS
   ```

2. Copy sample environment file:
   ```bash
   cp compose.example.env .env
   ```

3. Start container:
   ```bash
   docker compose up -d
   ```

4. Open `http://localhost:7700`.

---

## Configuration

HAMAL is configured via environment variables (prefixed with `HAMAL_*`, with `LAN_DROP_*` supported as fallback):

| Variable | Fallback | Default | Description |
| :--- | :--- | :--- | :--- |
| `HAMAL_LISTEN_ADDR` | `LAN_DROP_LISTEN_ADDR` | `:7700` | Address and port to bind HTTP server |
| `HAMAL_BASE_URL` | `LAN_DROP_BASE_URL` | *(empty)* | Optional base URL for QR codes and client links |
| `HAMAL_DATA_DIR` | `LAN_DROP_DATA_DIR` | `/data` | Root directory for SQLite DB and file storage |
| `HAMAL_DEFAULT_TTL` | `LAN_DROP_DEFAULT_TTL` | `1h` | Default room lifetime |
| `HAMAL_MIN_TTL` | `LAN_DROP_MIN_TTL` | `5m` | Minimum allowed room lifetime |
| `HAMAL_MAX_TTL` | `LAN_DROP_MAX_TTL` | `24h` | Maximum allowed room lifetime |
| `HAMAL_MAX_FILE_SIZE` | `LAN_DROP_MAX_FILE_SIZE` | `10GiB` | Maximum size of an individual file |
| `HAMAL_MAX_ROOM_SIZE` | `LAN_DROP_MAX_ROOM_SIZE` | `10GiB` | Maximum aggregate storage per room |
| `HAMAL_MAX_FILES_PER_ROOM` | `LAN_DROP_MAX_FILES_PER_ROOM` | `100` | Maximum number of files per room |
| `HAMAL_MAX_TOTAL_STORAGE` | `LAN_DROP_MAX_TOTAL_STORAGE` | `100GiB` | Global storage limit across all active rooms |
| `HAMAL_MIN_FREE_SPACE` | `LAN_DROP_MIN_FREE_SPACE` | `5GiB` | Minimum free disk space required on `/data` |
| `HAMAL_CLEANUP_INTERVAL` | `LAN_DROP_CLEANUP_INTERVAL` | `1m` | Sweep interval for expired rooms & orphans |
| `HAMAL_LOG_FORMAT` | `LAN_DROP_LOG_FORMAT` | `json` | Log format (`json` or `text`) |
| `HAMAL_LOG_LEVEL` | `LAN_DROP_LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `HAMAL_SECURE_COOKIES` | `LAN_DROP_SECURE_COOKIES` | `auto` | Cookie security (`auto`, `true`, `false`) |

---

## Brand Assets

Official HAMAL vector logos and icons are available in the [`brand/`](brand/) directory:

- `brand/hamal-logo-dark.svg` (Horizontal dark-mode wordmark)
- `brand/hamal-logo-light.svg` (Horizontal light-mode wordmark)
- `brand/hamal-logo-stacked.svg` (Stacked badge wordmark)
- `brand/hamal-app-icon.svg` (Square container icon)
- `brand/hamal-monochrome.svg` (Monochrome asset)
- `brand/hamal-path-mark.svg` (Vector path symbol)

---

## License

HAMAL is source-available software licensed under the
[HAMAL Source-Available Organizational Use License](LICENSE).

Organizations may freely use, modify, fork, and deploy HAMAL for their own
internal operations, including commercial business operations.

Commercial redistribution, resale, productization, and offering HAMAL or
derivative works as a commercial service require separate permission from
the copyright holder.
