# sshmgr

A macOS CLI tool to manage SSH targets on your local network.

---

## Requirements
- macOS
- Remote Login (SSH) enabled on target machines
- `ssh`, `dns-sd`, `pbcopy`, and `security` available (all built-in on macOS)
- Bonjour/mDNS enabled on the LAN for discovery

---

## Why sshmgr?
If you connect to Macs in the same LAN, IP addresses may change.  
`sshmgr` recommends using stable hostnames (e.g. `Mac-mini.local`). When the IP changes, you can still connect immediately because SSH uses the hostname, while sshmgr resolves and updates the last known IP for display.

---

## Bug Bounty

If you find a bug, email me at
sxiang36@outlook.com
(There is no bounty😭✌️)

---

## Features

- Host inventory: `add`, `list`, `show`, `rm`
- IP change hint: `check` and pre-SSH resolution
- One command to SSH: `sshmgr ssh <name>` (still connects via hostname)
- Passwords stored in **macOS Keychain**; copy to clipboard when needed
- Connection logs and stats: `history`, `users`
- LAN discovery via Bonjour `_ssh._tcp`: `discover`
- Enterprise-friendly filtering: `discover --probe` to classify `OK/AUTH/DENY/DOWN/ERR`
- Health check: `ping` (single host or all)

Discovery is based on Bonjour/mDNS service browsing of `_ssh._tcp` advertisements.

---

## Installation

Requires Go:

```
git clone https://github.com/stexiang/sshmgr.git
cd sshmgr
go mod tidy
go build -o sshmgr
```

---

## Storage 

- SQLite database (default):
  `~/.config/sshmgr/sshmgr.db`
- Passwords:
  Stored securely in macOS Keychain (never written to db)


## Quickstart

Add a Mac（Host recommend `.local`）

```
./sshmgr add <name> --user <user> --host <host>
```

Connect：

```
./sshmgr ssh <name>
```

Discover all hosts broadcasting SSH services and test their connectivity：

```
./sshmgr discover --probe --user <user> --only connectable
```

View statistics：

```
./sshmgr users
```

---

## Command Sheet

```
./sshmgr add <name> --user <user> --host <host>
./sshmgr list
./sshmgr show <name>
./sshmgr rm <name>
./sshmgr ssh <name> [--dry-run]
./sshmgr check <name>
```

Password

```
./sshmgr pass set <name>
./sshmgr pass copy <name> [--ttl 30]
./sshmgr pass clear <name>
```

History/Users

```
./sshmgr users
./sshmgr history [--name <name>] [--limit <n>]
```

Discover

```
./sshmgr discover [--probe] [--only connectable] [--add]
```

Ping

```
./sshmgr ping all [--timeout S] [--concurrency N] [--strict]
./sshmgr ping <name> [--timeout S] [--strict]
```

---

## License 

MIT

---

## Last

You are welcomed to create PRs😋

