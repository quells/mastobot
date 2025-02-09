package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	rootCmd.AddCommand(sqliteCmd)
}

var sqliteCmd = &cobra.Command{
	Use:   "sqlite",
	Short: "Test if SQLite was correctly cross-compiled",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		row := db.QueryRow(`WITH t AS (SELECT 1 AS c UNION SELECT 2 as C) SELECT SUM(c) FROM t`)
		var result int
		err = row.Scan(&result)
		if err != nil {
			return err
		}
		if result != 3 {
			return fmt.Errorf("expected 3, got %d", result)
		}
		fmt.Println("OK")
		return nil
	},
}
