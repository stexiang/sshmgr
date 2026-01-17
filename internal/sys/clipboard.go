package sys

import (
	"os/exec"
	"strconv"
	"strings"
)

func ClipboardCopy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func ClipboardClearAfter(ttlSeconds int) error {
	if ttlSeconds <= 0 {
		return nil
	}
	script := "sleep " + strconv.Itoa(ttlSeconds) + "; printf \"\" | pbcopy"
	cmd := exec.Command("sh", "-c", script)
	return cmd.Start()
}
