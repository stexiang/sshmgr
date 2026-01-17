package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"sshmgr/internal/sys"
)

var passTTL int

var passCmd = &cobra.Command{
	Use:   "pass",
	Short: "管理 Keychain 中保存的密码（只支持复制，不默认明文显示）",
}

var passSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "设置/更新密码到 Keychain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		var u string
		if err := DB.QueryRow(`SELECT user FROM hosts WHERE name=?`, name).Scan(&u); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Enter password for %s (%s): ", name, u)
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		pw := strings.TrimSpace(string(b))
		if pw == "" {
			return fmt.Errorf("empty password")
		}

		if err := sys.KeychainSet(name, u, pw); err != nil {
			return err
		}
		_, _ = DB.Exec(`UPDATE hosts SET has_secret=1 WHERE name=?`, name)
		return nil
	},
}

var passCopyCmd = &cobra.Command{
	Use:   "copy <name>",
	Short: "复制密码到剪贴板（可选 --ttl 自动清空）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		var u string
		if err := DB.QueryRow(`SELECT user FROM hosts WHERE name=?`, name).Scan(&u); err != nil {
			return err
		}

		pw, err := sys.KeychainGet(name, u)
		if err != nil {
			return err
		}
		if err := sys.ClipboardCopy(pw); err != nil {
			return err
		}
		_ = sys.ClipboardClearAfter(passTTL)
		fmt.Println("copied to clipboard")
		return nil
	},
}

var passClearCmd = &cobra.Command{
	Use:   "clear <name>",
	Short: "从 Keychain 删除密码",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		var u string
		if err := DB.QueryRow(`SELECT user FROM hosts WHERE name=?`, name).Scan(&u); err != nil {
			return err
		}
		if err := sys.KeychainDelete(name, u); err != nil {
			return err
		}
		_, _ = DB.Exec(`UPDATE hosts SET has_secret=0 WHERE name=?`, name)
		fmt.Println("deleted from keychain")
		return nil
	},
}

func init() {
	passCmd.AddCommand(passSetCmd, passCopyCmd, passClearCmd)
	passCopyCmd.Flags().IntVar(&passTTL, "ttl", 0, "clear clipboard after N seconds (0=disable)")
}
