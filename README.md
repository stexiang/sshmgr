# sshmgr

A macOS CLI to manage SSH targets on local networks.

`sshmgr` lets you connect with stable hostnames (for example `macmini.local`) instead of chasing changing IPs.

## Why sshmgr

In campus/office/home LANs, SSH workflows are often painful:

- DHCP changes IP addresses frequently
- Bonjour/mDNS may be partially blocked
- Password handling is annoying
- Connection history is hard to track

`sshmgr` keeps host records, resolves hostnames before connect, warns about IP changes, and stores passwords in macOS Keychain.

## Requirements

- macOS
- SSH enabled on target hosts (Remote Login)
- Built-in tools available:
  - `ssh`
  - `dns-sd`
  - `pbcopy`
  - `security`

## Installation

### One-line install (recommended)

```bash
curl -fsSL "https://raw.githubusercontent.com/stexiang/sshmgr/main/install.sh?ts=$(date +%s)" | bash
```

What the installer does:

- Tries matching GitHub Release assets first
- Falls back to source build automatically if no asset matches
- Installs to `/usr/local/bin` (or `~/.local/bin` if sudo is unavailable)

### Optional installer variables

```bash
SSHMGR_VERSION=v0.1.0 SSHMGR_INSTALL_DIR="$HOME/.local/bin" \
curl -fsSL "https://raw.githubusercontent.com/stexiang/sshmgr/main/install.sh?ts=$(date +%s)" | bash
```

- `SSHMGR_VERSION`: install a specific tag instead of `latest`
- `SSHMGR_INSTALL_DIR`: override install destination
- `SSHMGR_SKIP_RELEASE=1`: skip release probe and build from source directly
- `SSHMGR_NO_SPINNER=1`: disable spinner animation (useful in CI logs)

### Build from source manually

Requires Go.

```bash
git clone https://github.com/stexiang/sshmgr.git
cd sshmgr
go mod tidy
go build -o sshmgr .
./sshmgr --help
```

## Quick Start

### 1) Add a host

```bash
sshmgr add macmini --user yourname --host Mac-mini.local
```

### 2) Check resolution / reachability

```bash
sshmgr check macmini
sshmgr ping macmini
```

### 3) Connect

```bash
sshmgr ssh macmini
```

### 4) Save and copy password from Keychain

```bash
sshmgr pass set macmini
sshmgr pass copy macmini --ttl 30
```

## Discovery, Scan, and Reassociation

Discover hosts via Bonjour:

```bash
sshmgr discover
sshmgr discover --probe --user yourname --only connectable
```

Scan subnet when Bonjour is unreliable:

```bash
sshmgr scan 192.168.1.0/24
```

Reassociate a host after IP changes:

```bash
sshmgr reassociate macmini --subnet 192.168.1.0/24
```

## Command Overview

```text
add         Add a Mac target (recommended host: xxx.local)
check       Resolve host and report whether IP changed (updates last_ip)
discover    Discover SSH-enabled devices on the LAN (Bonjour: _ssh._tcp)
history     Show connection history (latest 20 by default)
list        List all host entries
pass        Manage passwords in Keychain (copy-only, no plaintext by default)
ping        Health check: resolve host and test TCP connectivity (default port 22)
reassociate Rediscover a host after IP change by scanning the subnet
rm          Remove one host entry (does not delete Keychain password)
scan        Scan subnet and detect SSH services
show        Show details of one host entry
ssh         Connect to target (resolves host, reports IP changes, writes history)
users       List entries as: name host ip count last pw
```

## Storage and Security

- Metadata DB: `~/.config/sshmgr/sshmgr.db`
- Passwords: macOS Keychain only
- Passwords are not written into SQLite

## Troubleshooting

### Help output still shows old text

You are likely running an old binary.

```bash
which sshmgr
sshmgr --version
```

If needed, rebuild/reinstall:

```bash
cd /path/to/sshmgr
go build -o sshmgr .
install -m 0755 sshmgr /usr/local/bin/sshmgr
```

### Installer appears stale

Use the default cache-busting URL (already included in this README):

```bash
curl -fsSL "https://raw.githubusercontent.com/stexiang/sshmgr/main/install.sh?ts=$(date +%s)" | bash
```

## License

MIT

## Contributing

Issues and PRs are welcome.
