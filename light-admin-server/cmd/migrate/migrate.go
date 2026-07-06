package migrate

import (
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
		if down {
			if err := db.Down(config.Database.PgxURL()); err != nil {
				logger.Zap.Fatalf("Error rolling back migration: %v", err)
			}
			logger.Zap.Info("Migration rolled back successfully")
			return
		}

		if err := db.Up(config.Database.PgxURL()); err != nil {
			logger.Zap.Fatalf("Error applying golang-migrate migrations: %v", err)
		}
		logger.Zap.Info("golang-migrate migrations applied successfully")

		// All modules have been migrated to sqlc: golang-migrate now owns every
		// table (through 000011), so there is no remaining GORM AutoMigrate step.
		logger.Zap.Info("Database migration completed successfully")
	},
}
