package cmd

import (
	"log"
	"os"

	"github.com/1995parham-teaching/redpanda101/internal/cmd/consumer"
	"github.com/1995parham-teaching/redpanda101/internal/cmd/migrate"
	"github.com/1995parham-teaching/redpanda101/internal/cmd/producer"
	"github.com/spf13/cobra"
)

// ExitFailure status code.
const ExitFailure = 1

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	//nolint: exhaustruct
	root := &cobra.Command{
		Use:   "redpanda 101",
		Short: "Use Redpanda instead of Kafka to process orders!",
	}

	root.PersistentFlags().StringP("config", "c", "config.toml", "path to config.toml")

	producer.Register(root)
	consumer.Register(root)
	migrate.Register(root)

	if err := root.Execute(); err != nil {
		log.Printf("failed to execute root command %s", err)
		os.Exit(ExitFailure)
	}
}
