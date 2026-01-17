package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	"sshmgr/internal/db"
	"sshmgr/internal/netx"
)

var (
	discoverTimeout int
	discoverAdd     bool
	discoverUser    string

	discoverProbe       bool
	discoverOnly        string
	discoverConcurrency int
	discoverProbeTO     int
)

type discItem struct {
	Instance string
	Domain   string
}

type discFound struct {
	Instance string
	Host     string
	Port     int
	IP       string
	Domain   string

	Status string // OK/AUTH/DENY/DOWN/ERR
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "发现局域网中开启 SSH 的设备（Bonjour: _ssh._tcp）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if discoverTimeout <= 0 {
			discoverTimeout = 3
		}
		if discoverConcurrency <= 0 {
			discoverConcurrency = 20
		}
		if discoverProbeTO <= 0 {
			discoverProbeTO = 2
		}
		if discoverOnly == "" {
			discoverOnly = "all"
		}

		// user：probe/add 都要用
		u := discoverUser
		if u == "" {
			if me, e := user.Current(); e == nil {
				u = me.Username
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(discoverTimeout)*time.Second)
		defer cancel()

		items, err := browseSSH(ctx, "local.")
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("no _ssh._tcp services found (try increasing --timeout)")
			return nil
		}

		found := make([]discFound, 0, len(items))
		for _, it := range items {
			h, p, err := lookupSSH(it)
			if err != nil {
				continue
			}
			ip := ""
			rctx, rcancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			if v, err := netx.ResolveHost(rctx, h); err == nil {
				ip = v
			}
			rcancel()

			found = append(found, discFound{
				Instance: it.Instance,
				Host:     h,
				Port:     p,
				IP:       ip,
				Domain:   it.Domain,
				Status:   "",
			})
		}

		// probe：并发探测
		if discoverProbe {
			if u == "" {
				return fmt.Errorf("cannot determine --user for probe, please pass --user")
			}
			probeAll(found, u, discoverProbeTO, discoverConcurrency)
		}

		// 过滤输出/添加
		filtered := make([]discFound, 0, len(found))
		for _, f := range found {
			if passOnlyFilter(f.Status, discoverOnly) {
				filtered = append(filtered, f)
			}
		}

		// 输出
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if discoverProbe {
			fmt.Fprintln(w, "NAME\tHOST\tPORT\tIP\tST")
			for _, f := range filtered {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", f.Instance, f.Host, f.Port, f.IP, f.Status)
			}
		} else {
			fmt.Fprintln(w, "NAME\tHOST\tPORT\tIP")
			for _, f := range filtered {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", f.Instance, f.Host, f.Port, f.IP)
			}
		}
		_ = w.Flush()

		// add：默认建议配合 probe，只加 connectable（你也可用 --only all 强制全加）
		if discoverAdd {
			if u == "" {
				return fmt.Errorf("cannot determine --user, please pass --user")
			}
			added := 0
			for _, f := range filtered {
				// 如果开了 probe 且你没手动指定 --only，这里默认只加 OK/AUTH（更符合你单位场景）
				if discoverProbe && discoverOnly == "all" {
					if f.Status != "OK" && f.Status != "AUTH" {
						continue
					}
				}

				name := uniqueName(slugify(preferNameFromHostOrInstance(f.Host, f.Instance)))
				_, _ = DB.Exec(`
INSERT INTO hosts(name,user,host,port,created_at)
VALUES(?,?,?,?,?)
ON CONFLICT(name) DO UPDATE SET
  user=excluded.user,
  host=excluded.host,
  port=excluded.port
`, name, u, f.Host, f.Port, db.NowUTC())
				added++
			}
			fmt.Printf("added/updated %d host(s) with user=%s\n", added, u)
		}

		return nil
	},
}

func init() {
	discoverCmd.Flags().IntVar(&discoverTimeout, "timeout", 3, "browse timeout seconds")
	discoverCmd.Flags().BoolVar(&discoverAdd, "add", false, "add discovered hosts into sshmgr db")
	discoverCmd.Flags().StringVar(&discoverUser, "user", "", "ssh user used with --add/--probe (default: current macOS user)")

	discoverCmd.Flags().BoolVar(&discoverProbe, "probe", false, "probe if the given user is connectable (OK/AUTH/DENY/DOWN)")
	discoverCmd.Flags().StringVar(&discoverOnly, "only", "all", "filter: all|connectable|ok|auth|deny|down|err")
	discoverCmd.Flags().IntVar(&discoverConcurrency, "concurrency", 20, "probe concurrency")
	discoverCmd.Flags().IntVar(&discoverProbeTO, "probe-timeout", 2, "probe timeout seconds per host")
}

// 解析 dns-sd browse 输出：收集 instance + domain
func browseSSH(ctx context.Context, domain string) ([]discItem, error) {
	c := exec.CommandContext(ctx, "dns-sd", "-B", "_ssh._tcp", domain)
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	c.Stderr = os.Stderr

	if err := c.Start(); err != nil {
		return nil, err
	}

	m := map[string]discItem{}
	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		// time Add flags if domain _ssh._tcp. <instance...>
		if len(fields) >= 7 && fields[1] == "Add" && fields[5] == "_ssh._tcp." {
			inst := strings.Join(fields[6:], " ")
			dom := fields[4]
			key := inst + "|" + dom
			m[key] = discItem{Instance: inst, Domain: dom}
		}
	}
	_ = c.Wait() // timeout/中止时可能非 0，忽略

	out := make([]discItem, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out, nil
}

// dns-sd -L "<instance>" _ssh._tcp <domain> 解析到 host:port
func lookupSSH(it discItem) (host string, port int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, "dns-sd", "-L", it.Instance, "_ssh._tcp", it.Domain)
	out, _ := c.CombinedOutput()

	re := regexp.MustCompile(`can be reached at\s+([^\s:]+):([0-9]+)`)
	m := re.FindStringSubmatch(string(out))
	if len(m) != 3 {
		return "", 0, fmt.Errorf("no SRV record for %s", it.Instance)
	}

	h := strings.TrimSuffix(m[1], ".")
	p := 0
	fmt.Sscanf(m[2], "%d", &p)
	if h == "" || p == 0 {
		return "", 0, fmt.Errorf("invalid SRV for %s", it.Instance)
	}
	return h, p, nil
}

// probe
func probeAll(found []discFound, user string, timeoutSeconds int, concurrency int) {
	type job struct {
		idx int
	}
	jobs := make(chan job)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			f := found[j.idx]
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
			st := probeOne(ctx, user, f.Host, f.Port)
			cancel()
			found[j.idx].Status = st
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker()
	}
	for i := range found {
		jobs <- job{idx: i}
	}
	close(jobs)
	wg.Wait()
}

func probeOne(ctx context.Context, user, host string, port int) string {
	target := fmt.Sprintf("%s@%s", user, host)
	args := []string{
		"-p", strconv.Itoa(port),
		"-o", "BatchMode=yes",
		"-o", "NumberOfPasswordPrompts=0",
		"-o", "ConnectTimeout=2",
		// probe 只为分类，不要卡 hostkey 交互，也不要污染 known_hosts
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		target,
		"exit",
	}

	c := exec.CommandContext(ctx, "ssh", args...)
	out, err := c.CombinedOutput()
	s := string(out)

	if err == nil {
		return "OK"
	}

	ls := strings.ToLower(s)

	// macOS/sshd 常见：User xxx not allowed because not listed in AllowUsers
	if strings.Contains(ls, "not allowed") || strings.Contains(ls, "allowusers") {
		return "DENY"
	}
	if strings.Contains(ls, "permission denied") {
		// 这里大概率是：允许该 user，但需要密码/密钥
		return "AUTH"
	}
	if strings.Contains(ls, "connection timed out") ||
		strings.Contains(ls, "operation timed out") ||
		strings.Contains(ls, "no route to host") ||
		strings.Contains(ls, "connection refused") {
		return "DOWN"
	}
	if strings.Contains(ls, "could not resolve hostname") ||
		strings.Contains(ls, "name or service not known") {
		return "ERR"
	}
	return "ERR"
}

func passOnlyFilter(status, only string) bool {
	switch only {
	case "all":
		return true
	case "connectable":
		return status == "OK" || status == "AUTH"
	case "ok":
		return status == "OK"
	case "auth":
		return status == "AUTH"
	case "deny":
		return status == "DENY"
	case "down":
		return status == "DOWN"
	case "err":
		return status == "ERR"
	default:
		return true
	}
}

func preferNameFromHostOrInstance(host, inst string) string {
	// host 常见是 XXX.local，优先用 XXX
	h := strings.TrimSuffix(host, ".local")
	h = strings.TrimSuffix(h, ".")
	if h != "" {
		return h
	}
	return inst
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if r == ' ' || r == '-' || r == '_' {
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	out = regexp.MustCompile(`-+`).ReplaceAllString(out, "-")
	if out == "" {
		out = "host"
	}
	return out
}

func uniqueName(base string) string {
	name := base
	for i := 2; ; i++ {
		var exists int
		_ = DB.QueryRow(`SELECT 1 FROM hosts WHERE name=? LIMIT 1`, name).Scan(&exists)
		if exists == 0 {
			return name
		}
		name = fmt.Sprintf("%s-%d", base, i)
	}
}
