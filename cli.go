package mirror

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type param struct {
	userAgent string
	outputDir string
}

var rootCmd = &cobra.Command{
	Use:   "mirror [url]",
	Short: "mirror is a command line tool for mirroring web page",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		url := args[0]
		userAgent, _ := cmd.Flags().GetString("agent")
		outputDir, _ := cmd.Flags().GetString("output-dir")

		param := param{
			userAgent: userAgent,
			outputDir: outputDir,
		}

		client := NewClient(param)
		client.mirror(url)
	},
}

func init() {
	rootCmd.Flags().StringP("agent", "A", "mirror/v0.0.1", "User-Agent name")
	rootCmd.Flags().StringP("output-dir", "o", "output", "Output Directory")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
