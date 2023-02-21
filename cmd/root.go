package cmd

import (
	"fmt"
	"os"
	"io/ioutil"

	"github.com/nicjohnson145/ksplit/config"
	"github.com/nicjohnson145/ksplit/internal"
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ksplit <file>",
		Short: "Split single-file sets of manifests into multiple files",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// So we don't print usage messages on execution errors
			cmd.SilenceUsage = true
			// So we dont double report errors
			cmd.SilenceErrors = true
			return config.InitializeConfig(cmd)
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var flBytes []byte
			var err error
			if args[0] == "-" {
				flBytes, err = ioutil.ReadAll(os.Stdin)
			} else {
				flBytes, err = os.ReadFile(args[0])
			}
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}

			split := internal.NewSplitter(flBytes)
			if err := split.Split(); err != nil {
				return fmt.Errorf("error splitting: %w", err)
			}
			return nil
		},
	}
	rootCmd.PersistentFlags().BoolP(config.Debug, "d", false, "Enable debug logging")

	return rootCmd
}
