package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/af-go/webrtc-demo/pkg/version"
	"github.com/spf13/cobra"
)

// VersionCmd version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		v := version.New()
		content, err := json.Marshal(v)
		if err != nil {
			fmt.Printf("Cannot get version information: %v\n", err)
			return
		}
		fmt.Printf("%s\n", string(content))
	},
}
