package cmd

import (
	"fmt"
	"os"

	"github.com/makkes/jardb/boltdb"
	"github.com/makkes/jardb/cmd/find"
	"github.com/makkes/jardb/cmd/index"
	"github.com/makkes/jardb/cmd/stats"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "jardb",
}

func Execute() {
	db := boltdb.NewBoltDB()
	if db == nil {
		os.Exit(1)
	}
	defer db.Close()
	rootCmd.AddCommand(find.NewCommand(db))
	rootCmd.AddCommand(index.NewCommand(db))
	rootCmd.AddCommand(stats.NewCommand(db))
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
