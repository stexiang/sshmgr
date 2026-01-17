package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有主机条目",
	RunE: func(cmd *cobra.Command, args []string) error {
		rows, err := DB.Query(`SELECT name,user,host,port,last_ip,last_checked_at,has_secret FROM hosts ORDER BY name`)
		if err != nil {
			return err
		}
		defer rows.Close()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tUSER\tHOST\tPORT\tLAST_IP\tLAST_CHECKED\tHAS_PASSWORD")

		for rows.Next() {
			var name, user, host string
			var port int
			var lastIP, lastChecked sql.NullString
			var hasSecret int

			if err := rows.Scan(&name, &user, &host, &port, &lastIP, &lastChecked, &hasSecret); err != nil {
				return err
			}

			checked := ""
			if lastChecked.Valid && lastChecked.String != "" {
				if t, e := time.Parse(time.RFC3339, lastChecked.String); e == nil {
					checked = t.Local().Format("2006-01-02 15:04:05")
				} else {
					checked = lastChecked.String
				}
			}

			has := "no"
			if hasSecret != 0 {
				has = "yes"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
				name, user, host, port, lastIP.String, checked, has)
		}

		_ = w.Flush()
		return rows.Err()
	},
}
