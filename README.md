# sshmgr

A macOS CLI tool to manage SSH targets on your local network.

`sshmgr` helps you connect to machines without caring about changing IP
addresses.
You keep using stable hostnames, and sshmgr takes care of the rest.

---

## What problem does it solve? 

In a local network, especially in school or enterprise environments:

- IP addresses change frequently
- Bonjour / mDNS may be blocked
- You forget which machine is which
- Password handling is annoying
- SSH history and usage are hard to track ：（

`sshmgr` is designed to make SSH simple and predictable again.

---

## Requirements

* macOS
* Remote Login (SSH) enabled on target machines
* Built-in macOS tools:

  * `ssh`
  * `dns-sd`
  * `pbcopy`
  * `security`

---

## Key Ideas

**Use hostnames, not IPs**

  - Recommended: `xxx.local`
**Always resolve before connecting**

  - IP changes are detected and recorded


---

## Features

- **Host inventory**

  - Add, list, show, and remove SSH targets
- **IP change detection**

  - Resolve hostname before connecting
  - Warn when IP changes and update records
- **One-command SSH**

  - `sshmgr ssh <name>`
  - Always connects via hostname
- **Password management**

  - Stored securely in macOS Keychain
  - Copy to clipboard when needed
- **History and statistics**

  - Track connection count and last access
- **LAN discovery**

  - Discover SSH-enabled machines via Bonjour (`_ssh._tcp`)
- **Active subnet scanning**

  - Scan a CIDR subnet to detect SSH services
  - Works in restricted networks where Bonjour is unavailable
- **Automatic reassociation**

  - Rediscover known hosts after IP changes by scanning the subnet
- **Health checks**

  - Quickly test SSH reachability

---

## Installation

Requires Go.

```bash
git clone https://github.com/stexiang/sshmgr.git
cd sshmgr
go mod tidy
go build -o sshmgr
```

---

## Storage

- **Metadata**

  - SQLite database:
    `~/.config/sshmgr/sshmgr.db`
- **Passwords**

  - Stored only in macOS Keychain
  - Never written to the database

---

## Quick Start

Add a machine
(recommended: use `.local` hostnames):

```bash
./sshmgr add macmini --user yourname --host Mac-mini.local
```

Connect:

```bash
./sshmgr ssh macmini
```

View usage statistics:

```bash
./sshmgr users
```

---

## Discovery

Discover machines broadcasting SSH via Bonjour:

```bash
./sshmgr discover
```

Filter connectable hosts only:

```bash
./sshmgr discover --probe --user yourname --only connectable
```

---

## Subnet Scan

In restricted networks where Bonjour does not work, you can actively scan a subnet:

```bash
./sshmgr scan 192.168.1.0/24
```

This detects hosts with SSH services without relying on mDNS or ARP.

---

## Automatic Reassociation

If a host changes its IP address and can no longer be reached:

```bash
./sshmgr reassociate macmini --subnet 192.168.1.0/24
```

`sshmgr` scans the subnet and verifies remote hostnames over SSH to rediscover the same machine, then updates the stored IP.

---

## Command Overview

Hosts:

```bash
./sshmgr add <name> --user <user> --host <host>
./sshmgr list
./sshmgr show <name>
./sshmgr rm <name>
```

Connect:

```bash
./sshmgr ssh <name> [--dry-run]
./sshmgr check <name>
```

Passwords:

```bash
./sshmgr pass set <name>
./sshmgr pass copy <name> [--ttl 30]
./sshmgr pass clear <name>
```

History and stats:

```bash
./sshmgr users
./sshmgr history [--name <name>] [--limit <n>]
```

Discovery and scanning:

```bash
./sshmgr discover [--probe] [--only connectable] [--add]
./sshmgr scan <subnet>
```

Reassociation:

```bash
./sshmgr reassociate <name> --subnet <cidr>
```

Health checks:

```bash
./sshmgr ping all
./sshmgr ping <name>
```

---

## License

MIT

---

## Contributing

Issues and pull requests are welcome😋
