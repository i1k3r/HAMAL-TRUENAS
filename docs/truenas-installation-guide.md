# TrueNAS SCALE Installation Guide for HAMAL

This guide explains how to deploy **HAMAL** on **TrueNAS SCALE** (24.10 "Electric Eel" or later) using the native Docker Compose **Install via YAML** feature.

---

## Overview & Security Architecture

HAMAL is a lightweight, secure, and private point-to-point local network file transfer service. On TrueNAS SCALE, HAMAL runs as an isolated, unprivileged container with the following security profile:

* **Non-Root Execution**: Runs strictly under UID `10001` and GID `10001` (`appuser`).
* **Read-Only Root Filesystem**: `read_only: true` locks down container binaries.
* **In-Memory Temporary Storage**: `tmpfs: /tmp` provides memory-backed volatile storage for temporary files.
* **Privilege Escalation Prevention**: `security_opt: [no-new-privileges:true]`.
* **Configurable Persistent Storage**: Database, room metadata, and temporary transfer chunks are stored in `/data` mounted to your designated TrueNAS ZFS dataset.

---

## Step 1: Create the ZFS Dataset

Before deploying the container, create a dedicated dataset on your storage pool:

1. In the TrueNAS SCALE WebUI, navigate to **Datasets**.
2. Select your target pool or parent dataset (e.g., `tank` or `tank/appdata`).
3. Click **Add Dataset**.
4. Enter **Dataset Name**: `hamal` (resulting dataset path: `/mnt/<pool_name>/appdata/hamal`).
5. Keep default settings (or select **Generic** / **Apps** preset) and click **Save**.

---

## Step 2: Configure Dataset Permissions (Crucial Step)

> [!CAUTION]
> **Creating a dataset is NOT enough.**
> By default, newly created TrueNAS datasets are owned by `root:root` with restricted access. Because HAMAL runs as an unprivileged, non-root user (`UID 10001`, `GID 10001`), the host dataset mounted to `/data` **must explicitly grant write access to UID 10001 and GID 10001**.
> If permissions are not configured, the container will fail on startup with `failed to open database: permission denied`.

### Recommended Method: TrueNAS WebUI Permissions / ACL Editor

1. In the TrueNAS WebUI, go to **Datasets** and select your `hamal` dataset.
2. In the **Permissions** card on the right, click **Edit** (or **Set ACL** / **Edit ACL**).
3. **If using standard POSIX Permissions**:
   * Set **User**: `10001` and check **Apply User**.
   * Set **Group**: `10001` and check **Apply Group**.
   * Ensure **Read**, **Write**, and **Execute** checkboxes are checked for User and Group.
4. **If using an NFSv4 / POSIX ACL**:
   * Click **Add Item** -> Who: **User** -> User: `10001` -> Permissions: **Full Control** (or Read + Write + Execute).
   * Click **Add Item** -> Who: **Group** -> Group: `10001` -> Permissions: **Full Control** (or Read + Write + Execute).
5. Check **Apply permissions recursively** (and **Apply permissions to child datasets** if applicable).
6. Click **Save Access Control List** / **Save**.

*(Optional CLI Alternative via SSH / TrueNAS Shell)*:
```bash
chown -R 10001:10001 /mnt/<pool_name>/appdata/hamal
chmod -R 770 /mnt/<pool_name>/appdata/hamal
```

---

## Step 3: Deploy via TrueNAS WebUI ("Install via YAML")

1. In the TrueNAS SCALE WebUI, navigate to **Apps**.
2. Click **Discover Apps** in the upper right.
3. Click the **(⋮)** menu icon in the upper right corner and select **Install via YAML**.
4. Configure the deployment:
   * **Application Name**: `hamal`
   * **Docker Compose YAML**: Paste the configuration from [`truenas/compose.yaml`](../truenas/compose.yaml):

```yaml
services:
  hamal:
    image: ghcr.io/i1k3r/hamal-truenas:latest
    container_name: hamal
    restart: unless-stopped

    # Hardened unprivileged security
    read_only: true
    user: "10001:10001"
    security_opt:
      - no-new-privileges:true

    ports:
      - "7700:7700"

    volumes:
      # Replace /mnt/YOUR_POOL/appdata/hamal with your actual TrueNAS dataset path
      - /mnt/YOUR_POOL/appdata/hamal:/data

    tmpfs:
      - /tmp:size=64M,mode=1777

    environment:
      - TZ=Europe/Istanbul
      # - HAMAL_BASE_URL=http://192.168.1.100:7700 # Optional: explicit LAN URL for QR codes
      # - HAMAL_MAX_FILE_SIZE=10GiB                # Optional: max file upload limit
      # - HAMAL_DEFAULT_TTL=1h                     # Optional: default room lifespan
```

> [!NOTE]
> Replace `/mnt/YOUR_POOL/appdata/hamal` with the exact path of your dataset created in Step 1 (for example, `/mnt/tank/appdata/hamal`).

5. Click **Install** / **Save** to start the container.

---

## Step 4: Access HAMAL & Verify Status

1. **Web Interface Access**:
   Open a browser and navigate to:
   ```text
   http://<TRUENAS_IP>:7700
   ```
2. **Verify Application Logs**:
   In the TrueNAS **Apps** list, verify that `hamal` displays status **RUNNING**. Click on the application card to inspect the logs. A normal startup will display:
   ```json
   {"time":"...","level":"INFO","msg":"server starting","listen_addr":":7700"}
   ```
3. **Container Healthcheck**:
   The Docker image includes a built-in healthcheck (`/app/hamal healthcheck`) that validates application readiness.

---

## Step 5: Troubleshooting

### Symptom: `failed to open database: permission denied` or exit status 1

* **Cause**: The mounted host dataset `/data` does not grant write access to UID `10001` or GID `10001`.
* **Resolution**:
  1. Go to **Datasets** -> select `hamal` dataset -> **Edit Permissions**.
  2. Confirm UID `10001` and GID `10001` have Read and Write access.
  3. Check **Apply permissions recursively** and save.
  4. In **Apps**, restart `hamal`.

### Port Conflict (Port 7700 Already in Use)

If port `7700` is used by another service on TrueNAS, adjust the host port mapping:
```yaml
    ports:
      - "7701:7700"  # Exposes HAMAL on host port 7701
```
