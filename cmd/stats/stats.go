package stats

import (
	"fmt"

	"github.com/makkes/jardb/db"
	"github.com/spf13/cobra"
)

func NewCommand(db db.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show stats about the index",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			stats := db.Stats()
			fmt.Printf("Classes: %d\nJars: %d\n", stats.ClassCount, stats.JarCount)
		},
	}
}
