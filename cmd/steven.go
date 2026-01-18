package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const stevenURL = "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=RDdQw4w9WgXcQ&start_radio=1"

var stevenCmd = &cobra.Command{
	Use:    "steven",
	Short:  "hidden easter egg",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 打印 banner
		fmt.Print(banner)
		fmt.Println("Warning, The action you are doing right now is DANGEROUS! It might cause a serious damage to your computer! (y/N)")

		// 仅在交互终端允许继续
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("not a TTY: refusing to run interactive easter egg")
		}

		ok, err := askYesNo("Continue?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("You saved your mac.")
			return nil
		}

		// 必须输入 proceed
		fmt.Println()
		fmt.Println("Second warning: type 'proceed' to fuck your computer, anything else to keep your mac alive.")
		if err := requireExactInput("proceed"); err != nil {
			fmt.Println("You saved your mac.")
			return nil
		}

		// 打开网页（macOS）
		return openURL(stevenURL)
	},
}

func askYesNo(prompt string) (bool, error) {
	fmt.Printf("%s ", prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}
	s := strings.TrimSpace(strings.ToLower(line))
	return s == "y" || s == "yes", nil
}

func requireExactInput(expected string) error {
	fmt.Print("> ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(line) != expected {
		return fmt.Errorf("input mismatch")
	}
	return nil
}

func openURL(url string) error {
	// macOS: open <url>
	c := exec.Command("open", url)
	return c.Start()
}
