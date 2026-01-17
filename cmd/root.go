package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"

	"sshmgr/internal/app"
	"sshmgr/internal/db"
)

// 注意：banner 这里不能包含反引号 ` ，否则会把 Go 的 raw string 截断
const banner = `
              __
   __________/ /_  ____ ___  ____ ______
  / ___/ ___/ __ \/ __ '__ \/ __ '/ ___/
 (__  |__  ) / / / / / / / / /_/ / /
/____/____/_/ /_/_/ /_/ /_/\__, /_/
                          /____/

`

var (
	version = "1.0"
	author  = "Steven, 2026"
)

var (
	dbPath string
	DB     *sql.DB
)

var rootCmd = &cobra.Command{
	Use:   "sshmgr",
	Short: "管理局域网 Mac 的 SSH 条目、Keychain 密码，并提示 IP 变更",
	Long:  "sshmgr：管理 SSH 目标（推荐 host 为 xxx.local），连接前解析并提示 IP 变化；密码存 Keychain，仅支持复制。\n",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if DB != nil {
			return nil
		}
		if dbPath == "" {
			dbPath = app.DefaultDBPath()
		}
		if err := app.EnsureParentDir(dbPath); err != nil {
			return err
		}

		d, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return err
		}
		if err := db.Init(d); err != nil {
			_ = d.Close()
			return err
		}
		DB = d
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", app.DefaultDBPath(), "sqlite db path")

	// help 顶部显示 banner
	tpl := banner + `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:
  {{.UseLine}}

{{if .HasAvailableSubCommands}}Commands:
{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "help"))}}  {{rpad .Name .NamePadding}} {{.Short}}
{{end}}{{end}}{{end}}

{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
{{if .HasAvailableInheritedFlags}}Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`
	rootCmd.SetHelpTemplate(tpl)

	// --version 输出：version + author（带 banner）
	rootCmd.Version = fmt.Sprintf("%s\nversion=%s\nauthor=%s\n", banner, version, author)
	rootCmd.SetVersionTemplate("{{.Version}}")

	rootCmd.AddCommand(addCmd, listCmd, showCmd, rmCmd, checkCmd, sshCmd, usersCmd, historyCmd, passCmd, discoverCmd, pingCmd)
}
