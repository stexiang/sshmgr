package cmd

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"sshmgr/internal/netutil"
	"sshmgr/internal/sshutil"
)

var (
	scanTimeout     time.Duration
	scanConcurrency int
	scanUser        string
)

func init() {
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 500*time.Millisecond, "dial timeout")
	scanCmd.Flags().IntVar(&scanConcurrency, "concurrency", 64, "number of workers")
	scanCmd.Flags().StringVar(&scanUser, "user", "", "user for SSH hostname fallback")
	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan <subnet>",
	Short: "Scan subnet and detect SSH services",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ips, err := expandSubnet(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Scanning %d hosts ...\n", len(ips))

		sem := make(chan struct{}, scanConcurrency)
		out := make(chan string)

		var wg sync.WaitGroup

		for _, ip := range ips {
			ip := ip
			wg.Add(1)

			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// 1. SSH banner probe
				banner, err := netutil.SSHBanner(ip, scanTimeout)
				if err != nil || banner == "" {
					return
				}

				hostname := netutil.ExtractHostname(banner)

				// 2. reverse DNS fallback
				if hostname == "" {
					if ptr, _ := netutil.ReverseLookup(ip); ptr != "" {
						hostname = ptr
					}
				}

				// 3. remote hostname via ssh
				if hostname == "" && scanUser != "" {
					if h, _ := sshutil.RemoteHostname(scanUser, ip); h != "" {
						hostname = h
					}
				}

				if hostname == "" {
					hostname = "(unknown)"
				}

				out <- fmt.Sprintf("%s  %s", ip, hostname)
			}()
		}

		go func() {
			wg.Wait()
			close(out)
		}()

		found := 0
		for line := range out {
			fmt.Println(line)
			found++
		}

		fmt.Printf("\nFound %d SSH host(s)\n", found)
		return nil
	},
}

// expandSubnet expands CIDR like 10.203.9.0/24 into usable host IPs.
func expandSubnet(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	base := binary.BigEndian.Uint32(ipnet.IP.To4())
	mask := binary.BigEndian.Uint32(net.IP(ipnet.Mask).To4())

	start := base & mask
	end := start | ^mask

	var ips []string
	for i := start + 1; i < end; i++ {
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], i)
		ips = append(ips, net.IP(b[:]).String())
	}

	return ips, nil
}
