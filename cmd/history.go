package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var (
	histName  string
	histLimit int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "查看连接历史（默认最近 20 条）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if histLimit <= 0 {
			histLimit = 20
		}

		rows, err := DB.Query(`
SELECT h.name,h.user,h.host,c.resolved_ip,c.end_at,c.duration_ms,c.exit_code
FROM conn_log c
JOIN hosts h ON h.id=c.host_id
WHERE (?='' OR h.name=?)
ORDER BY c.id DESC
LIMIT ?`, histName, histName, histLimit)
		if err != nil {
			return err
		}
		defer rows.Close()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tUSER\tHOST\tIP\tEND_AT\tDURATION_MS\tEXIT")

		for rows.Next() {
			var name, user, host string
			var ip, endAt sql.NullString
			var dur int64
			var exit int

			if err := rows.Scan(&name, &user, &host, &ip, &endAt, &dur, &exit); err != nil {
				return err
			}

			end := endAt.String
			if endAt.Valid && endAt.String != "" {
				if t, e := time.Parse(time.RFC3339, endAt.String); e == nil {
					end = t.Local().Format("2006-01-02 15:04:05")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\n", name, user, host, ip.String, end, dur, exit)
		}

		_ = w.Flush()
		return rows.Err()
	},
}

func init() {
	historyCmd.Flags().StringVar(&histName, "name", "", "filter by host name")
	historyCmd.Flags().IntVar(&histLimit, "limit", 20, "max rows")
}
