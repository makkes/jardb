package find

import (
	"fmt"
	"github.com/makkes/jardb/db"
	"github.com/spf13/cobra"
)

func NewCommand(db db.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "find <package or class name>",
		Short: "Find a class or package in the DB",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for res := range db.Find(args[0]) {
				fmt.Println(res)
			}
		},
	}
}
