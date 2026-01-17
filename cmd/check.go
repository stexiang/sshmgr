package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sshmgr/internal/db"
	"sshmgr/internal/netx"
)

var checkCmd = &cobra.Command{
	Use:   "check <name>",
	Short: "解析 host 并提示 IP 是否变化（更新 last_ip）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var (
			id     int64
			host   string
			lastIP sql.NullString
		)
		err := DB.QueryRow(`SELECT id, host, last_ip FROM hosts WHERE name=?`, name).Scan(&id, &host, &lastIP)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ip, err := netx.ResolveHost(ctx, host)
		if err != nil {
			return err
		}
		if ip == "" {
			return fmt.Errorf("no IP resolved for host: %s", host)
		}

		if lastIP.Valid && lastIP.String != "" && lastIP.String != ip {
			fmt.Printf("IP changed: %s -> %s\n", lastIP.String, ip)
		} else {
			fmt.Printf("IP: %s\n", ip)
		}

		_, err = DB.Exec(`UPDATE hosts SET last_ip=?, last_checked_at=? WHERE id=?`, ip, db.NowUTC(), id)
		return err
	},
}
