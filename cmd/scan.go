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
	Short: "Scan subnet, detect SSH hosts, extract names and fingerprint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ips, err := expandSubnet(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Scanning %d hosts ...\n", len(ips))

		ch := make(chan string)
		var wg sync.WaitGroup
		sem := make(chan struct{}, scanConcurrency)

		for _, ip := range ips {
			ip := ip
			wg.Add(1)

			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// Step 1: banner probe
				banner, berr := netutil.SSHBanner(ip, scanTimeout)
				if berr != nil || banner == "" {
					return
				}

				hostname := netutil.ExtractHostname(banner)

				// Step 2: reverse DNS fallback
				if hostname == "" {
					if ptr, _ := netutil.ReverseLookup(ip); ptr != "" {
						hostname = ptr
					}
				}

				// Step 3: remote hostname via SSH
				if hostname == "" && scanUser != "" {
					if host, _ := sshutil.RemoteHostname(scanUser, ip); host != "" {
						hostname = host
					}
				}

				// Step 4: fingerprint identity
				fprint, _ := sshutil.Fingerprint(ip)

				if hostname == "" {
					hostname = "(unknown)"
				}

				ch <- fmt.Sprintf("%s  %s  %s", ip, hostname, fprint[:12]) // short fingerprint
			}()
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		found := 0
		for line := range ch {
			fmt.Println(line)
			found++
		}

		fmt.Printf("\nFound %d SSH host(s)\n", found)
		return nil
	},
}
// expandSubnet takes CIDR like 10.203.9.0/24 and returns all host IPs.
func expandSubnet(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	var ips []string

	// network:   first IP
	// broadcast: last IP
	network := binary.BigEndian.Uint32(ipnet.IP.To4())
	mask := binary.BigEndian.Uint32(net.IP(ipnet.Mask).To4())
	start := network & mask
	end := start | ^mask

	// iterate usable host range (skip network + broadcast)
	for ip := start + 1; ip < end; ip++ {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, ip)
		ips = append(ips, net.IP(b).String())
	}

	return ips, nil
}
