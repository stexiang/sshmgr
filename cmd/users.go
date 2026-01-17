package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "按条目列出：name host ip count last pw",
	RunE: func(cmd *cobra.Command, args []string) error {
		rows, err := DB.Query(`
SELECT
  h.name,
  h.user,
  h.host,
  h.last_ip,
  h.has_secret,
  (SELECT COUNT(1) FROM conn_log c WHERE c.host_id = h.id) AS conn_count,
  (SELECT c.end_at FROM conn_log c WHERE c.host_id = h.id ORDER BY c.id DESC LIMIT 1) AS last_connected_at,
  (SELECT c.resolved_ip FROM conn_log c WHERE c.host_id = h.id ORDER BY c.id DESC LIMIT 1) AS last_connected_ip
FROM hosts h
ORDER BY COALESCE(last_connected_at, '') DESC, h.name ASC;
`)
		if err != nil {
			return err
		}
		defer rows.Close()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tHOST\tIP\tCOUNT\tLAST\tPW")

		for rows.Next() {
			var (
				name       string
				userIgnore string
				host       string

				lastIP     sql.NullString
				hasSecret  int
				connCount  int
				lastConnAt sql.NullString
				lastConnIP sql.NullString
			)

			if err := rows.Scan(&name, &userIgnore, &host, &lastIP, &hasSecret, &connCount, &lastConnAt, &lastConnIP); err != nil {
				return err
			}

			ip := ""
			if lastConnIP.Valid && lastConnIP.String != "" {
				ip = lastConnIP.String
			} else if lastIP.Valid {
				ip = lastIP.String
			}

			last := ""
			if lastConnAt.Valid && lastConnAt.String != "" {
				if t, e := time.Parse(time.RFC3339, lastConnAt.String); e == nil {
					last = t.Local().Format("01-02 15:04")
				} else {
					last = lastConnAt.String
				}
			}

			pw := "no"
			if hasSecret != 0 {
				pw = "yes"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", name, host, ip, connCount, last, pw)
		}

		_ = w.Flush()
		return rows.Err()
	},
}
