# sshmgr

macOS 下的命令行 SSH 管理工具
**局域网里有一堆 Mac？忘 IP，记不住密码？用这个直接 ssh。**

---

## Bug Bounty

如果你在使用过程中找到了漏洞，请提交到sxiang36@outlook.com。（没有奖赏😂）

---

## 为啥造这个工具

学校里很多 Mac：

* IP 一天一个样
* 每次连都要找密码
* 想知道最近连哪个最多也麻烦
* 最主要还是想连同学整蛊

于是我做了这个工具，从 只支持 add/ssh，一路加到现在这样

---

## Features / 能干啥

| 功能                 | 解释                              |
| ------------------ | ------------------------------- |
| `add/list/show/rm` | 自己维护一份 SSH CMDB                 |
| 自动解析 hostname      | IP 变了也能直接连                      |
| Keychain 管密码       | 随时复制密码                  |
| `ssh <name>`       | 一条命令就连                          |
| 记历史                | 每次连接结束时间，时长，出口吗                 |
| `users` 统计         | 哪台机子最常连，一眼看全局                   |
| Discover           | Bonjour 探测 `_ssh._tcp` 找局域网的新机器 |
| Probe              | 过滤出“能连 / 要密码 / 拒绝 / 挂了 / 报错”    |
| Ping all           | 批量检查 ssh 端口，顺便更新 last_ip        |
| SQLite 存库          | 默认 `~/.config/sshmgr/sshmgr.db` |

**不用记 IP，不用找密码，也不用找设备。**

---


## Install / 安装

依赖 Go：

```
git clone <your repo>
cd sshmgr
go mod tidy
go build -o sshmgr
```

---

## Quickstart / 快速开始

添加一台 Mac（推荐 `.local`）

```
./sshmgr add <name> --user <user> --host <host>
```

一键连：

```
./sshmgr ssh <name>
```

找所有广播 ssh 的机器：

```
./sshmgr discover --add --user <your_user>
```

查看统计：

```
./sshmgr users
```

---

## 常用命令

```
./sshmgr add <name> --user <user> --host <host>
./sshmgr list
./sshmgr show <name>
./sshmgr rm <name>
./sshmgr ssh <name> [--dry-run]
./sshmgr check <name>
```

密码

```
./sshmgr pass set <name>
./sshmgr pass copy <name> [--ttl 30]
./sshmgr pass clear <name>
```

历史

```
./sshmgr users
./sshmgr history [--name <name>] [--limit <n>]
```

发现

```
./sshmgr discover [--probe] [--only connectable] [--add]
```

探测

```
./sshmgr ping all [--timeout S] [--concurrency N] [--strict]
./sshmgr ping <name> [--timeout S] [--strict]
```

---

## License 

MIT

---

## 最后

没最后了，欢迎加PR。

