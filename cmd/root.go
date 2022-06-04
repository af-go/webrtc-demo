package cmd

import (
	"fmt"
	"os"

	"github.com/af-go/webrtc-demo/cmd/analyze"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "webrtc",
	Short: "",
}

func init() {
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(analyze.AnalyzeCmd)
}

// Exec excute root command
func Exec() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
