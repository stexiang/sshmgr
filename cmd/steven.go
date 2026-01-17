// cmd/steven.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const stevenBanner = `
  _________ __                           
 /   _____//  |_   ____ ___  __ ____   ____  
 \_____  \\   __\_/ __ \\  \/ // __ \ /    \ 
 /        \|  |  \  ___/ \   /\  ___/|   |  \
/_______  /|__|   \___  > \_/  \___  >___|  /
        \/            \/           \/     \/ 


`

var stevenCmd = &cobra.Command{
	Use:    "steven",
	Short:  "hidden easter egg",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(stevenBanner)
	},
}
