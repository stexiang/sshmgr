package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"sshmgr/internal/db"
	"sshmgr/internal/netx"
)

var (
	pingTimeout     int
	pingConcurrency int
	pingStrict      bool
)

type pingRow struct {
	ID   int64
	Name string
	Host string
	Port int
}

type pingResult struct {
	Name string
	Host string
	IP   string
	Port int
	MS   int64
	ST   string // OK/DOWN/RESOLVE/ERR
	Err  error
}

var pingCmd = &cobra.Command{
	Use:   "ping <name|all>",
	Short: "健康检查：解析 host 并测试 TCP 连接到端口（默认 22）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pingTimeout <= 0 {
			pingTimeout = 2
		}
		if pingConcurrency <= 0 {
			pingConcurrency = 30
		}

		target := args[0]
		var rows []pingRow
		var err error

		if target == "all" {
			rows, err = loadAllHosts()
		} else {
			rows, err = loadOneHost(target)
		}
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			fmt.Println("no hosts")
			return nil
		}

		results := runPing(rows, time.Duration(pingTimeout)*time.Second, pingConcurrency)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tHOST\tIP\tPORT\tMS\tST")
		fail := 0
		for _, r := range results {
			if r.ST != "OK" {
				fail++
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n", r.Name, r.Host, r.IP, r.Port, r.MS, r.ST)
		}
		_ = w.Flush()

		if pingStrict && fail > 0 {
			return fmt.Errorf("%d host(s) not healthy", fail)
		}
		return nil
	},
}

func init() {
	pingCmd.Flags().IntVar(&pingTimeout, "timeout", 2, "timeout seconds per host")
	pingCmd.Flags().IntVar(&pingConcurrency, "concurrency", 30, "concurrency for ping all")
	pingCmd.Flags().BoolVar(&pingStrict, "strict", false, "exit non-zero if any host is not OK")
}

func loadAllHosts() ([]pingRow, error) {
	rs, err := DB.Query(`SELECT id,name,host,port FROM hosts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var out []pingRow
	for rs.Next() {
		var r pingRow
		if err := rs.Scan(&r.ID, &r.Name, &r.Host, &r.Port); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rs.Err()
}

func loadOneHost(name string) ([]pingRow, error) {
	var r pingRow
	err := DB.QueryRow(`SELECT id,name,host,port FROM hosts WHERE name=?`, name).Scan(&r.ID, &r.Name, &r.Host, &r.Port)
	if err != nil {
		return nil, err
	}
	return []pingRow{r}, nil
}

func runPing(rows []pingRow, timeout time.Duration, conc int) []pingResult {
	type job struct {
		idx int
	}
	jobs := make(chan job)
	results := make([]pingResult, len(rows))

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for j := range jobs {
			row := rows[j.idx]
			results[j.idx] = pingOne(row, timeout)
		}
	}

	for i := 0; i < conc; i++ {
		wg.Add(1)
		go worker()
	}
	for i := range rows {
		jobs <- job{idx: i}
	}
	close(jobs)
	wg.Wait()
	return results
}

func pingOne(h pingRow, timeout time.Duration) pingResult {
	res := pingResult{Name: h.Name, Host: h.Host, Port: h.Port, ST: "ERR"}

	// resolve
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ip, err := netx.ResolveHost(ctx, h.Host)
	cancel()
	if err != nil || ip == "" {
		res.ST = "RESOLVE"
		res.Err = err
		return res
	}
	res.IP = ip

	// 顺便更新 last_ip/last_checked_at
	_, _ = DB.Exec(`UPDATE hosts SET last_ip=?, last_checked_at=? WHERE id=?`, ip, db.NowUTC(), h.ID)

	// 2) tcp connect
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", h.Port))
	d := net.Dialer{Timeout: timeout}

	start := time.Now()
	c, err := d.Dial("tcp", addr)
	ms := time.Since(start).Milliseconds()
	res.MS = ms

	if err != nil {
		res.ST = "DOWN"
		res.Err = err
		return res
	}
	_ = c.Close()
	res.ST = "OK"
	return res
}

// ensure we keep sql imported (some environments complain if unused by build tags)
var _ sql.NullString
