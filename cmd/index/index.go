package index

import (
	"github.com/makkes/jardb/db"
	"github.com/spf13/cobra"
)

func NewCommand(db db.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "index <folder>",
		Short: "Add all JARs in a folder to the index",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			db.IndexFolders(args[0])
		},
	}
}
