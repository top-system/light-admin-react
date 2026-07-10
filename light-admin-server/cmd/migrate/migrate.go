package migrate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/top-system/light-admin/db"
	"github.com/top-system/light-admin/lib"
)

var (
	configFile string
	down       bool
)

func init() {
	pf := StartCmd.PersistentFlags()
	pf.StringVarP(&configFile, "config", "c",
		"config/config.yaml", "this parameter is used to start the service application")
	pf.BoolVar(&down, "down", false, "roll back the most recent sqlc/golang-migrate migration and exit")
}

var StartCmd = &cobra.Command{
	Use:          "migrate",
	Short:        "Migrate database",
	Example:      "{execfile} migrate -c config/config.yaml",
	SilenceUsage: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		lib.SetConfigPath(configFile)
	},
	Run: func(cmd *cobra.Command, args []string) {
		config := lib.NewConfig()
		logger := lib.NewLogger(config)

		// golang-migrate owns the sqlc-managed tables (t_user, t_user_role, ...).
		// The engine picks the migration dialect directory; the URL scheme picks
		// the migrate driver.
		engine, url := config.Database.EngineName(), config.Database.MigrateURL()
		if down {
			if err := db.Down(engine, url); err != nil {
				logger.Error(fmt.Sprintf("Error rolling back migration: %v", err))
				os.Exit(1)
			}
			logger.Info("Migration rolled back successfully")
			return
		}

		if err := db.Up(engine, url); err != nil {
			logger.Error(fmt.Sprintf("Error applying golang-migrate migrations: %v", err))
			os.Exit(1)
		}
		logger.Info("golang-migrate migrations applied successfully")

		// All modules have been migrated to sqlc: golang-migrate now owns every
		// table (through 000011), so there is no remaining GORM AutoMigrate step.
		logger.Info("Database migration completed successfully")
	},
}
