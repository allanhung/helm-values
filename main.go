package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const desc = `
Merge one or more YAML files of values.
	$ helm values mychart -f path/to/merging/file

To write to a file, instead of stdout, use '-o':
	$ helm values mychart -f path/to/merging/file -o path/to/output/dir/
`

func main() {
	valuesCmd := valuesCmd{}

	cmd := &cobra.Command{
		Use:   "helm values [flags] CHART",
		Short: fmt.Sprintf("merge one or more YAML files of values"),
		Long:  desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("chart is required")
			}
			// verify chart path exists
			if _, err := os.Stat(args[0]); err == nil {
				if valuesCmd.chartPath, err = filepath.Abs(args[0]); err != nil {
					return err
				}
			} else {
				return err
			}
			// verify that output-dir exists if provided
			if valuesCmd.outputDir != "" {
				_, err := os.Stat(valuesCmd.outputDir)
				if os.IsNotExist(err) {
					return fmt.Errorf("output-dir '%s' does not exist", valuesCmd.outputDir)
				}
			}
			return valuesCmd.run()
		},
	}
	f := cmd.Flags()
	f.VarP(&valuesCmd.values, "values", "f", "specify values in a YAML file (can specify multiple)")
	f.StringVarP(&valuesCmd.valTemplate, "value-template", "t", "", "specify value template in a YAML file")
	f.VarP(&valuesCmd.sections, "section", "s", "specify section in values template file (can specify multiple)")
	f.StringVarP(&valuesCmd.outputDir, "output-dir", "o", "", "writes the merged values to files in output-dir instead of stdout")
	f.StringVarP(&valuesCmd.backupSuffix, "backup-suffix", "", ".bak", "suffix append to values.yaml if values.yaml already exist in output-dir")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
