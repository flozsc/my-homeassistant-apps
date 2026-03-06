# Home Assistant App Development – Vibe Coding Reference

> Built from the official source docs at https://github.com/home-assistant/developers.home-assistant/tree/master/docs/apps
> Covers: configuration, communication, presentation, publishing, repository, security, testing, tutorial

---

## What are Apps?

Apps (formerly "add-ons") extend Home Assistant by running containerised services alongside it — things like an MQTT broker, Samba share, or any custom service. They are managed via **Settings → Apps** in the HA UI, powered by the Supervisor.

Under the hood every app is a Docker container image, published to a registry (GitHub Container Registry or Docker Hub) or built locally on the user's device.

**Key repos:**
- [Template / example app repo](https://github.com/home-assistant/addons-example)
- [Official HA core apps](https://github.com/home-assistant/addons)
- [Community apps](https://github.com/hassio-addons)
- [Docker base images](https://github.com/home-assistant/docker-base)
- [Builder tool](https://github.com/home-assistant/builder)

---

## File Structure

Each app lives in its own folder. A repository can hold multiple apps.

```
my-repo/
├── repository.yaml          # required at repo root
├── my_app/
│   ├── config.yaml          # app configuration (required)
│   ├── Dockerfile           # container definition (required)
│   ├── run.sh               # startup script
│   ├── build.yaml           # optional: custom base images / build args
│   ├── CHANGELOG.md
│   ├── DOCS.md              # user-facing documentation
│   ├── README.md            # shown in the app store
│   ├── icon.png             # 128x128px square PNG
│   ├── logo.png             # ~250x100px PNG
│   ├── apparmor.txt         # optional: custom AppArmor profile
│   └── translations/
│       └── en.yaml          # optional: UI label translations
```

> `config`, `build`, and translation files all accept `.json`, `.yml`, or `.yaml`.

---

## `config.yaml` – Full Reference

### Required fields

| Key | Type | Description |
|-----|------|-------------|
| `name` | string | Display name of the app |
| `version` | string | Semver version. Must match the image tag if using `image`. Bump to trigger update prompt. |
| `slug` | string | Unique identifier within the repository. Must be URI-friendly. |
| `description` | string | Short description |
| `arch` | list | Supported architectures: `armhf`, `armv7`, `aarch64`, `amd64`, `i386` |

### Optional fields

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `machine` | list | all | Restrict to specific machine types. Prefix with `!` to negate. |
| `url` | url | | Homepage / support thread for the app |
| `startup` | string | `application` | Startup order: `initialize` (during HA setup), `system` (databases etc.), `services` (before HA), `application` (after HA), `once` (non-daemon) |
| `webui` | string | | URL for the app's web UI, e.g. `http://[HOST]:[PORT:2839]/dashboard`. Port is internal and gets substituted. Protocol can also bind to a config option: `[PROTO:option_name]://[HOST]:[PORT:2839]/dashboard` |
| `boot` | string | `auto` | `auto` = system-controlled, `manual` = user-started only, `manual_only` = never auto-start (user cannot change) |
| `ports` | dict | | Ports to expose: `"container-port/type": host-port`. Set host port to `null` to disable. |
| `ports_description` | dict | | Human-readable port descriptions: `"container-port/type": "description"` |
| `host_network` | bool | `false` | Run on host network. Limits addressability by other apps. |
| `host_ipc` | bool | `false` | Share IPC namespace with host |
| `host_dbus` | bool | `false` | Map host D-Bus into the app |
| `host_pid` | bool | `false` | Run in host PID namespace. Unprotected apps only. **Warning:** incompatible with S6 Overlay. |
| `host_uts` | bool | `false` | Use host UTS namespace |
| `devices` | list | | Host devices to map in, e.g. `/dev/ttyAMA0` |
| `homeassistant` | string | | Minimum required HA Core version, e.g. `2022.10.5` |
| `hassio_role` | string | `default` | Supervisor API role: `default`, `homeassistant`, `backup`, `manager`, `admin` |
| `hassio_api` | bool | `false` | Access to Supervisor REST API at `http://supervisor/` |
| `homeassistant_api` | bool | `false` | Access to HA Core REST API proxy at `http://supervisor/core/api` |
| `auth_api` | bool | `false` | Access to HA user auth backend |
| `docker_api` | bool | `false` | Read-only Docker API access. Unprotected apps only. |
| `privileged` | list | | Linux capabilities: `BPF`, `CHECKPOINT_RESTORE`, `DAC_READ_SEARCH`, `IPC_LOCK`, `NET_ADMIN`, `NET_RAW`, `PERFMON`, `SYS_ADMIN`, `SYS_MODULE`, `SYS_NICE`, `SYS_PTRACE`, `SYS_RAWIO`, `SYS_RESOURCE`, `SYS_TIME` |
| `full_access` | bool | `false` | Full hardware access like Docker privileged mode. Unprotected apps only. Do not combine with `devices`, `uart`, `usb`, or `gpio`. |
| `apparmor` | bool/string | `true` | Enable/disable AppArmor, or specify a custom profile name |
| `map` | list | | HA directory mounts (see below) |
| `environment` | dict | | Extra environment variables |
| `audio` | bool | `false` | Map PulseAudio into container |
| `video` | bool | `false` | Map all available video devices |
| `gpio` | bool | `false` | Map `/sys/class/gpio`. May also need `/dev/mem` and `SYS_RAWIO`. Disable or customise AppArmor if enabled. |
| `usb` | bool | `false` | Map `/dev/bus/usb` with plug & play support |
| `uart` | bool | `false` | Auto-map all host UART/serial devices |
| `udev` | bool | `false` | Mount host udev database read-only |
| `devicetree` | bool | `false` | Map `/device-tree` |
| `kernel_modules` | bool | `false` | Map host kernel modules + config (read-only), grants `SYS_MODULE` |
| `stdin` | bool | `false` | Enable STDIN via HA API |
| `legacy` | bool | `false` | Legacy mode for images without `hass.io` labels |
| `options` | dict | | Default values for user-configurable options |
| `schema` | dict | | Validation schema for options. Set to `false` to disable validation entirely. |
| `image` | string | | Container image name, e.g. `ghcr.io/home-assistant/{arch}-addon-example`. Use `{arch}` for multi-arch. |
| `codenotary` | string | | Email address for Codenotary CAS image signing |
| `timeout` | integer | `10` | Seconds to wait for Docker daemon before killing |
| `tmpfs` | bool | `false` | Mount `/tmp` as tmpfs (memory filesystem) |
| `discovery` | list | | Services this app provides to HA |
| `services` | list | | Services consumed/provided. Format: `service:function` where function is `provide`, `want`, or `need`. Supported services: `mqtt`, `mysql` |
| `ingress` | bool | `false` | Enable Ingress (proxy web UI through HA frontend) |
| `ingress_port` | integer | `8099` | Port your server listens on. Set `0` for host-network apps (read via API). |
| `ingress_entry` | string | `/` | URL entry point |
| `ingress_stream` | bool | `false` | Stream requests to the app |
| `panel_icon` | string | `mdi:puzzle` | MDI icon for sidebar panel |
| `panel_title` | string | | Sidebar title (defaults to app name) |
| `panel_admin` | bool | `true` | Restrict panel to admin users |
| `backup` | string | `hot` | `hot` = backup while running, `cold` = Supervisor stops app first (ignores `backup_pre`/`backup_post`) |
| `backup_pre` | string | | Command to run before backup |
| `backup_post` | string | | Command to run after backup |
| `backup_exclude` | list | | Files/paths (glob supported) to exclude from backups |
| `advanced` | bool | `false` | Only show to users with Advanced mode enabled |
| `stage` | string | `stable` | `stable`, `experimental`, or `deprecated`. Non-stable hidden unless user has Advanced mode. |
| `init` | bool | `true` | Set `false` if the image has its own init system. S6 Overlay v3+ **requires** `false`. |
| `watchdog` | string | | Health check URL. Same format as `webui`. TCP: `tcp://[HOST]:[PORT:80]`. Works for host and internal network. |
| `realtime` | bool | `false` | Access to host scheduler + `SYS_NICE` |
| `journald` | bool | `false` | Mount host system journal read-only. Check `/var/log/journal`, fallback `/run/log/journal`. |
| `breaking_versions` | list | | Versions requiring manual update — auto-update skipped if update would cross these |
| `ulimits` | dict | | Resource limits. Each key is a limit name; value is an integer or `{soft: N, hard: N}`. Must not exceed host hard limit. |

### `map` — directory mounts

```yaml
map:
  - type: homeassistant_config   # mounted at /homeassistant in container
    read_only: false
    path: /custom/config/path    # optional: override mount path inside container
  - type: addon_config           # /config — user-provided config files (read-only)
  - type: addon_config:rw        # /config — read-write
  - type: ssl                    # /ssl
  - type: share                  # /share
  - type: media                  # /media
  - type: addons                 # /addons
  - type: backup                 # /backup
  - type: all_addon_configs      # all addon config directories
  - type: data                   # /data — always mapped + writable; path can be customised
```

`/data` is always mounted and writable. `/data/options.json` contains the user-set options.

### `options` / `schema`

`options` sets defaults; `schema` validates user input. Set a default to `null` to make the option mandatory. Use `?` suffix in schema (and omit from `options`) to make it truly optional (no default).

```yaml
options:
  log_level: "info"
  mqtt_host: null        # mandatory — user must fill in before app can start

schema:
  log_level: "list(debug|info|warning|error)"
  mqtt_host: str
  port: "int(1,65535)"
  enabled: bool
  api_key: "str?"        # optional, no entry in options
  endpoint: url
  pattern: "match(^\\w+$)"
  ratio: "float(0,1)"
  device: "device(subsystem=tty)"
```

**Supported schema types:**

| Type | Notes |
|------|-------|
| `str` | Also `str(min,)` / `str(,max)` / `str(min,max)` |
| `bool` | |
| `int` | Also `int(min,)` / `int(,max)` / `int(min,max)` |
| `float` | Also `float(min,)` / `float(,max)` / `float(min,max)` |
| `email` | |
| `url` | |
| `password` | |
| `port` | |
| `match(REGEX)` | |
| `list(a\|b\|c)` | Enum / dropdown |
| `device` | `device(subsystem=tty)` for serial devices |

Nested arrays and dicts supported to a **maximum depth of 2**.

### Removing deprecated config keys

```bash
options=$(bashio::addon.options)
old_key='my_old_option'
if bashio::jq.exists "${options}" ".${old_key}"; then
    bashio::log.info "Removing ${old_key}"
    bashio::addon.option "${old_key}"   # no second argument = delete
fi
```

---

## Dockerfile

All apps are based on Alpine Linux. Use HA base images for automatic arch substitution.

```dockerfile
ARG BUILD_FROM
FROM $BUILD_FROM

# Install dependencies
RUN \
  apk add --no-cache \
    python3

WORKDIR /data

COPY run.sh /
RUN chmod a+x /run.sh

CMD [ "/run.sh" ]
```

If **not** using the HA build system, add labels manually:

```dockerfile
LABEL \
  io.hass.version="VERSION" \
  io.hass.type="addon" \
  io.hass.arch="armhf|aarch64|i386|amd64"
```

Architecture-specific Dockerfiles are supported: `Dockerfile.amd64`, `Dockerfile.aarch64`, etc.

**Build args available by default:**

| ARG | Description |
|-----|-------------|
| `BUILD_FROM` | Base image (substituted automatically) |
| `BUILD_VERSION` | App version from `config.yaml` |
| `BUILD_ARCH` | Current build architecture |

---

## `run.sh` – Startup Script

```bash
#!/usr/bin/with-contenv bashio

# Read a single option from /data/options.json
TARGET=$(bashio::config 'target')

# Read from the services API (e.g. MQTT)
MQTT_HOST=$(bashio::services mqtt "host")
MQTT_USER=$(bashio::services mqtt "username")
MQTT_PASSWORD=$(bashio::services mqtt "password")

bashio::log.info "Starting with target: ${TARGET}"

exec my-service --target "${TARGET}"
```

Options are also available directly as `/data/options.json`.

> Use UNIX line endings (LF) — not Windows CRLF.

---

## `build.yaml` – Extended Build Options

Only needed for custom base images or additional build args.

```yaml
build_from:
  armhf:   mycustom/base-image:latest
  aarch64: mycustom/base-image:latest
  amd64:   mycustom/base-image:latest
args:
  my_build_arg: "value"
labels:
  my.custom.label: "value"
codenotary:
  signer: you@example.com
  base_image: notary@home-assistant.io   # verifies official HA base images
```

| Key | Required | Description |
|-----|----------|-------------|
| `build_from` | no | Dict of `arch: base-image` |
| `args` | no | Additional Docker build arguments |
| `labels` | no | Additional Docker labels |
| `codenotary.signer` | no | Signer email for Codenotary CAS |
| `codenotary.base_image` | no | Base image to verify. Use `notary@home-assistant.io` for official HA images. |

---

## Communication

### Internal network

Apps communicate over an internal network using DNS names in the format `{REPO}_{SLUG}` (replace `_` with `-` for a valid hostname).

- Locally installed: `local_my_app` → hostname `local-my-app`
- From GitHub repo: `3283fh_my_app` → hostname `3283fh-my-app`

Use `supervisor` as the hostname to reach the Supervisor. Apps on `host_network: true` can address all internal apps by name, but cannot themselves be addressed by name from other apps (alias still works).

### Home Assistant Core REST API

```yaml
# config.yaml
homeassistant_api: true
```

```bash
curl -X GET \
  -H "Authorization: Bearer ${SUPERVISOR_TOKEN}" \
  -H "Content-Type: application/json" \
  http://supervisor/core/api/config
```

### Home Assistant WebSocket API

```
ws://supervisor/core/websocket
# Use SUPERVISOR_TOKEN as the password
```

### Supervisor API

```yaml
# config.yaml
hassio_api: true
hassio_role: default   # or homeassistant / backup / manager / admin
```

```bash
curl -H "Authorization: Bearer ${SUPERVISOR_TOKEN}" http://supervisor/
```

**Endpoints accessible without `hassio_api: true`:**
- `/core/api`, `/core/api/stream`, `/core/websocket`
- `/addons/self/*`
- `/services*`, `/discovery*`, `/info`

### Services API (inter-app)

```yaml
# config.yaml
services:
  - mqtt:need      # provide / want / need
  - mysql:want
```

```bash
MQTT_HOST=$(bashio::services mqtt "host")
MQTT_USER=$(bashio::services mqtt "username")
MQTT_PASSWORD=$(bashio::services mqtt "password")
```

### STDIN

Send data from HA automations via the `hassio.addon_stdin` action. Requires `stdin: true` in `config.yaml`.

---

## Ingress (Embedded Web UI)

Ingress proxies your app's web UI through the HA frontend with HA handling authentication. Earns +2 security points, no port management needed.

```yaml
# config.yaml
ingress: true
ingress_port: 8099
ingress_entry: /
ingress_stream: false   # set true for streaming/websocket-heavy UIs
```

**Requirements for your server:**
- Listen on `ingress_port` (default 8099)
- Accept connections **only** from `172.30.32.2` — deny all others
- No authentication needed
- Use header `X-Ingress-Path` if you need to know your base URL

**Supported protocols:** HTTP/1.x, streaming, WebSockets

**User identity headers (sent by Supervisor on every request):**

| Header | Description |
|--------|-------------|
| `X-Remote-User-Id` | HA user ID |
| `X-Remote-User-Name` | Username |
| `X-Remote-User-Display-Name` | Display name |

**Minimal Nginx example:**

`ingress.conf`:
```nginx
server {
    listen 8099;
    allow  172.30.32.2;
    deny   all;
}
```

`Dockerfile`:
```dockerfile
ARG BUILD_FROM
FROM $BUILD_FROM

RUN apk --no-cache add nginx && mkdir -p /run/nginx

COPY ingress.conf /etc/nginx/http.d/

CMD [ "nginx", "-g", "daemon off;error_log /dev/stdout debug;" ]
```

`config.yaml`:
```yaml
name: "Ingress Example"
version: "1.0.0"
slug: "nginx-ingress-example"
description: "Ingress testing"
arch:
  - amd64
  - armhf
  - armv7
  - i386
ingress: true
```

---

## Translations

File: `translations/en.yaml`

```yaml
configuration:
  log_level:
    name: "Log Level"
    description: "Verbosity of logging"
  ssh:
    name: "SSH Options"
    description: "Configure SSH authentication"
    fields:
      public_key:
        name: "Public Key"
        description: "Client public key"
network:
  8080/tcp: "Web interface port"
```

Keys under `configuration` must match `schema` keys. Keys under `network` must match `ports` keys.

---

## User-Provided Config Files (`addon_config`)

Use when users need to supply config files directly to your internal service:

```yaml
# config.yaml
map:
  - type: addon_config      # read-only, mounted at /config
  # or:
  - type: addon_config:rw   # read-write
```

Users place files in: `/addon_configs/{REPO}_<slug>/` (locally: `local_<slug>`)

Mounted at `/config` inside the container. Also useful for exposing app output (logs, databases, generated files) back to the user.

---

## Security

Apps start at **rating 5/6**. Protection mode is on by default.

### Rating changes

| Action | Change | Notes |
|--------|--------|-------|
| `ingress: true` | +2 | Overrides `auth_api` |
| `auth_api: true` | +1 | Overridden by `ingress` |
| Signed with Codenotary CAS | +1 | |
| Custom `apparmor.txt` | +1 | Applied after installation |
| `apparmor: false` | -1 | |
| `privileged: NET_ADMIN/SYS_ADMIN/SYS_RAWIO/SYS_PTRACE/SYS_MODULE/DAC_READ_SEARCH` or `kernel_modules` | -1 | Once even if multiple used |
| `hassio_role: manager` | -1 | |
| `host_network: true` | -1 | |
| `host_uts: true` + `SYS_ADMIN` | -1 | |
| `hassio_role: admin` | -2 | |
| `host_pid: true` | -2 | |
| `full_access: true` | → 1 | Overrides everything |
| `docker_api: true` | → 1 | Overrides everything |

### Supervisor API roles

| Role | Access |
|------|--------|
| `default` | All `info` endpoints |
| `homeassistant` | All HA API endpoints |
| `backup` | All backup API endpoints |
| `manager` | Extended rights (for CLIs) |
| `admin` | Full access, can toggle protection mode |

### Best practices

- Avoid `host_network` unless necessary
- Write a custom `apparmor.txt`
- Mount directories read-only unless write is needed
- Request only the API permissions you actually use
- Sign images with Codenotary CAS
- Use `auth_api: true` + HA Auth backend instead of plain-text credentials in options

### AppArmor profile template

Place `apparmor.txt` alongside `config.yaml`. Replace `ADDON_SLUG` with your slug.

```
#include <tunables/global>

profile ADDON_SLUG flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  file,
  signal (send) set=(kill,term,int,hup,cont),

  # S6-Overlay
  /init ix,
  /bin/** ix,
  /usr/bin/** ix,
  /run/{s6,s6-rc*,service}/** ix,
  /package/** ix,
  /command/** ix,
  /etc/services.d/** rwix,
  /etc/cont-init.d/** rwix,
  /etc/cont-finish.d/** rwix,
  /run/{,**} rwk,
  /dev/tty rw,

  # Bashio
  /usr/lib/bashio/** ix,
  /tmp/** rwk,

  # Persistent data
  /data/** rw,

  # Start new profile for your service binary
  /usr/bin/myprogram cx -> myprogram,

  profile myprogram flags=(attach_disconnected,mediate_deleted) {
    #include <abstractions/base>
    signal (receive) peer=*_ADDON_SLUG,
    /data/** rw,
    /share/** rw,
    /usr/bin/myprogram r,
    /bin/bash rix,
    /bin/echo ix,
    /etc/passwd r,
    /dev/tty rw,
  }
}
```

**Workflow for writing a profile:**
1. Add minimum known access
2. Add `complain` flag to the profile
3. Run the app; audit with `journalctl _TRANSPORT="audit" -g 'apparmor="ALLOWED"'`
4. Add access until no audit warnings remain
5. Remove `complain` so ungranted access is DENIED

---

## Repository Configuration

`repository.yaml` at the repo root is required:

```yaml
name: "My App Repository"
url: "https://github.com/yourname/my-ha-apps"
maintainer: "Your Name <you@example.com>"
```

| Key | Required | Description |
|-----|----------|-------------|
| `name` | yes | Repository display name |
| `url` | no | Homepage |
| `maintainer` | no | Contact info |

**Installing:** Settings → Apps → Store icon → paste URL → Save.
One-click install button generator: https://my.home-assistant.io/create-link/

**Canary/beta branch:** append `#branchname` to the repo URL:
```
https://github.com/yourname/my-ha-apps#next
```
Consider different `repository.yaml` names per branch, e.g. "My App (stable)" vs "My App (beta)".

---

## Publishing

### Pre-built images (preferred)

```yaml
# config.yaml
image: "ghcr.io/yourname/{arch}-my-app"
```

`{arch}` is substituted at install time. `version` in `config.yaml` must match the image tag.

**Branch strategy:** `main`/`master` = latest published tag. Develop on a `build` branch or PR; merge after pushing to registry.

**Build using the HA Builder — from a git repo:**

```bash
docker run \
  --rm --privileged \
  -v ~/.docker/config.json:/root/.docker/config.json:ro \
  ghcr.io/home-assistant/amd64-builder \
  --all \
  -t addon-folder \
  -r https://github.com/yourname/my-ha-apps \
  -b branchname
```

**Build from a local directory:**

```bash
docker run \
  --rm --privileged \
  -v ~/.docker/config.json:/root/.docker/config.json:ro \
  -v /my_addon:/data \
  ghcr.io/home-assistant/amd64-builder \
  --all \
  -t /data
```

### Locally built containers

No image registry needed — Supervisor builds on the user's device. Good for prototyping, bad for production (slow install, SD card wear, dependency risk). Comment out `image` to force local build:

```yaml
#image: ghcr.io/yourname/{arch}-my-app
```

---

## Local Testing

### VS Code devcontainer (recommended)

1. Install [Remote Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. Copy [`devcontainer.json`](https://github.com/home-assistant/devcontainer/raw/main/addons/devcontainer.json) → `.devcontainer/devcontainer.json`
3. Copy [`tasks.json`](https://github.com/home-assistant/devcontainer/raw/main/addons/tasks.json) → `.vscode/tasks.json`
4. Open folder in VS Code → "Reopen in Container"
5. Terminal → Run Task → **Start Home Assistant**
6. HA available at `http://localhost:7123/`
7. Your app folder is automatically a Local App

Works on Windows, Mac, Linux.

### Remote development (physical hardware)

Install Samba or SSH app on a real HA device, copy your app folder to `/addons/<my_app>/`. Useful for apps that need real serial ports or hardware.

### Local build with Docker (no devcontainer)

Build for all architectures:

```bash
docker run \
  --rm -it --name builder --privileged \
  -v /path/to/addon:/data \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  ghcr.io/home-assistant/amd64-builder \
  -t /data --all --test \
  -i my-test-addon-{arch} \
  -d local
```

Build single arch with standalone Docker:

```bash
docker build \
  --build-arg BUILD_FROM="ghcr.io/home-assistant/amd64-base:latest" \
  -t local/my-test-addon \
  .
```

Base images: `ghcr.io/home-assistant/{armhf,aarch64,amd64,i386}-base:latest`

Run locally:

```bash
docker run --rm \
  -v /tmp/my_test_data:/data \
  -p 8000:8000 \
  local/my-test-addon
```

### Debugging

App not appearing after "Check for updates"? `config.yaml` is invalid. Check: Settings → System → Logs → **Supervisor** dropdown. Validation errors at the bottom.

All `stdout`/`stderr` goes to Docker logs, visible in the app's Logs tab in the Supervisor panel.

---

## Presentation

| File | Spec |
|------|------|
| `README.md` | Short intro shown in the app store |
| `DOCS.md` | Full user-facing documentation |
| `CHANGELOG.md` | Version history (follow [keepachangelog.com](https://keepachangelog.com)) |
| `logo.png` | ~250×100px PNG |
| `icon.png` | 128×128px square PNG |

---

## Key Environment Variables

| Variable | Description | Requires |
|----------|-------------|----------|
| `SUPERVISOR_TOKEN` | Bearer token for HA Core and Supervisor APIs | `homeassistant_api: true` or `hassio_api: true` |
| `TZ` | Timezone | Add `tzdata` to Dockerfile if needed |

---

## Quick-Start Checklist

```
[ ] my_app/config.yaml           name, version, slug, arch, options, schema
[ ] my_app/Dockerfile            FROM $BUILD_FROM, install deps, CMD
[ ] my_app/run.sh                read options via bashio, start service (LF line endings!)
[ ] repository.yaml              at repo root
[ ] my_app/DOCS.md               explain all config options
[ ] my_app/translations/en.yaml  label options in the UI
[ ] my_app/logo.png              ~250x100px
[ ] my_app/icon.png              128x128px square
[ ] my_app/apparmor.txt          custom profile (+1 security point)
```
