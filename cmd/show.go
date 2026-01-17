package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "查看某条主机详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var user, host, note, tags, lastIP, lastChecked, created string
		var port, hasSecret int

		err := DB.QueryRow(`
SELECT user,host,port,note,tags,last_ip,last_checked_at,has_secret,created_at
FROM hosts WHERE name=?`, name).Scan(
			&user, &host, &port, &note, &tags, &lastIP, &lastChecked, &hasSecret, &created,
		)
		if err != nil {
			return err
		}

		fmt.Printf(
			"name: %s\nuser: %s\nhost: %s\nport: %d\nnote: %s\ntags: %s\nlast_ip: %s\nlast_checked_at: %s\nhas_password: %v\ncreated_at: %s\n",
			name, user, host, port, note, tags, lastIP, lastChecked, hasSecret != 0, created,
		)
		return nil
	},
}
