package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"sshmgr/internal/db"
	"sshmgr/internal/netutil"
	"sshmgr/internal/sshutil"
)

var (
	reSubnet      string
	reTimeout     time.Duration
	reConcurrency int
)

func init() {
	reassociateCmd.Flags().StringVar(&reSubnet, "subnet", "", "CIDR subnet to scan (required)")
	reassociateCmd.Flags().DurationVar(&reTimeout, "timeout", 800*time.Millisecond, "dial timeout")
	reassociateCmd.Flags().IntVar(&reConcurrency, "concurrency", 32, "scan concurrency")
	_ = reassociateCmd.MarkFlagRequired("subnet")

	rootCmd.AddCommand(reassociateCmd)
}

var reassociateCmd = &cobra.Command{
	Use:   "reassociate <name>",
	Short: "Rediscover a host after IP change by scanning the subnet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// 1. load host record
		var (
			id   int64
			user string
			host string
		)

		err := DB.QueryRow(
			`SELECT id, user, host FROM hosts WHERE name=?`,
			name,
		).Scan(&id, &user, &host)
		if err != nil {
			return fmt.Errorf("host %q not found", name)
		}

		fmt.Printf("Reassociating %s (%s)...\n", name, host)

		// 2. expand subnet
		ips, err := expandSubnet(reSubnet)
		if err != nil {
			return err
		}

		sem := make(chan struct{}, reConcurrency)
		found := make(chan string, 1)

		var wg sync.WaitGroup

		for _, ip := range ips {
			ip := ip
			wg.Add(1)

			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// must have SSH
				if _, err := netutil.SSHBanner(ip, reTimeout); err != nil {
					return
				}

				// authoritative check: remote hostname
				h, err := sshutil.RemoteHostname(user, ip)
				if err != nil {
					return
				}

				if h == host {
					select {
					case found <- ip:
					default:
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(found)
		}()

		ip, ok := <-found
		if !ok {
			fmt.Println("No matching host found")
			return nil
		}

		// 3. update last_ip
		_, _ = DB.Exec(
			`UPDATE hosts SET last_ip=?, last_checked_at=? WHERE id=?`,
			ip,
			db.NowUTC(),
			id,
		)

		fmt.Printf("Reassociated %s -> %s\n", name, ip)
		return nil
	},
}
