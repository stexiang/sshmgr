package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/spf13/cobra"
	"sshmgr/internal/db"
	"sshmgr/internal/netx"
)

var sshDryRun bool

var sshCmd = &cobra.Command{
	Use:   "ssh <name>",
	Short: "连接目标（连接前解析并提示 IP 变化，同时写入连接历史）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var (
			id     int64
			u      string
			host   string
			port   int
			lastIP sql.NullString
		)
		if err := DB.QueryRow(`SELECT id,user,host,port,last_ip FROM hosts WHERE name=?`, name).
			Scan(&id, &u, &host, &port, &lastIP); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		ip, err := netx.ResolveHost(ctx, host)
		if err == nil && ip != "" {
			// 提示变更 + 更新 last_ip
			if lastIP.Valid && lastIP.String != "" && lastIP.String != ip {
				fmt.Printf("IP changed: %s -> %s\n", lastIP.String, ip)
			}
			_, _ = DB.Exec(`UPDATE hosts SET last_ip=?, last_checked_at=? WHERE id=?`, ip, db.NowUTC(), id)
		}

		target := fmt.Sprintf("%s@%s", u, host)
		argsSSH := []string{"-p", fmt.Sprintf("%d", port), target}

		if sshDryRun {
			fmt.Printf("dry-run: ssh %v\n", argsSSH)
			return nil
		}

		start := time.Now()
		c := exec.Command("ssh", argsSSH...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		err = c.Run()
		end := time.Now()

		exitCode := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exitCode = ee.ProcessState.ExitCode()
			} else {
				exitCode = -1
			}
		}

		localUser := ""
		if me, e := user.Current(); e == nil {
			localUser = me.Username
		}

		_, _ = DB.Exec(`
INSERT INTO conn_log(host_id,start_at,end_at,duration_ms,resolved_ip,exit_code,local_user)
VALUES(?,?,?,?,?,?,?)
`, id,
			start.UTC().Format(time.RFC3339),
			end.UTC().Format(time.RFC3339),
			end.Sub(start).Milliseconds(),
			ip,
			exitCode,
			localUser,
		)

		return err
	},
}

func init() {
	sshCmd.Flags().BoolVar(&sshDryRun, "dry-run", false, "print ssh command without executing")
}
