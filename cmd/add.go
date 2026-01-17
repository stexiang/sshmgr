package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"sshmgr/internal/db"
)

var (
	addUser string
	addHost string
	addPort int
	addNote string
	addTags string
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "添加一台 Mac 目标（建议 host 用 xxx.local）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if addUser == "" || addHost == "" {
			return fmt.Errorf("--user and --host are required")
		}
		if addPort <= 0 || addPort > 65535 {
			return fmt.Errorf("invalid --port: %s", strconv.Itoa(addPort))
		}

		_, err := DB.Exec(`
INSERT INTO hosts(name,user,host,port,note,tags,created_at)
VALUES(?,?,?,?,?,?,?)
ON CONFLICT(name) DO UPDATE SET
  user=excluded.user,
  host=excluded.host,
  port=excluded.port,
  note=excluded.note,
  tags=excluded.tags
`, name, addUser, addHost, addPort, addNote, addTags, db.NowUTC())
		return err
	},
}

func init() {
	addCmd.Flags().StringVar(&addUser, "user", "", "ssh user (required)")
	addCmd.Flags().StringVar(&addHost, "host", "", "host or hostname (required, e.g. Mac-mini.local)")
	addCmd.Flags().IntVar(&addPort, "port", 22, "ssh port")
	addCmd.Flags().StringVar(&addNote, "note", "", "note")
	addCmd.Flags().StringVar(&addTags, "tags", "", "tags (comma-separated)")
}
