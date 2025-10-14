// Package cmd provides command-line interface functionality for the RSSFFS application.
//
// This package implements the root command and manages the command-line interface
// using the cobra library. It handles RSS feed discovery and subscription operations,
// logging setup, and command execution for the RSSFFS (RSS Feed Finder and Subscriber)
// application.
//
// The package integrates with several components:
//   - Configuration management through pkg/config
//   - Core RSS functionality through internal/RSSFFS
//   - Manual pages through pkg/man
//   - Version information through pkg/version
//
// Example usage:
//
//	import "github.com/toozej/RSSFFS/cmd/RSSFFS"
//
//	func main() {
//		cmd.Execute()
//	}
package cmd

import (
	"fmt"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/toozej/RSSFFS/internal/RSSFFS"
	"github.com/toozej/RSSFFS/pkg/config"
	"github.com/toozej/RSSFFS/pkg/man"
	"github.com/toozej/RSSFFS/pkg/version"
)

// Command-line flag variables that store user input from CLI flags.
//
// These variables are populated by cobra flag parsing and used throughout
// the application execution. They control various aspects of RSSFFS behavior.
var (
	// category specifies the RSS reader category name to assign new feeds to.
	// Set via the --category/-c flag.
	category string

	// debug enables debug-level logging when set to true.
	// Set via the --debug/-d flag.
	debug bool

	// clearCategoryFeeds determines whether to delete all existing feeds
	// within the specified category before subscribing to new feeds.
	// Set via the --clearCategoryFeeds/-r flag.
	clearCategoryFeeds bool
)

// rootCmd defines the base command for the RSSFFS CLI application.
//
// This command serves as the entry point for RSS feed discovery and subscription
// operations. It accepts a single URL argument and processes it to find and
// subscribe to RSS feeds found on that page and linked pages.
//
// Command characteristics:
//   - Requires exactly one URL argument
//   - Supports persistent flags inherited by subcommands
//   - Validates URL format before processing
//   - Integrates with RSS reader API for feed subscription
var rootCmd = &cobra.Command{
	Use:              "RSSFFS [pageURL]",
	Short:            "RSS Feed Finder [and] Subscriber",
	Long:             `Automatically find and subscribe to RSS feeds found on inputted URL, and on URLs mentioned on the inputted URL.`,
	Args:             cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	PersistentPreRun: rootCmdPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		conf := config.GetEnvVars()

		inputURL := args[0]
		pageURL, err := url.ParseRequestURI(inputURL)
		if err != nil {
			fmt.Println("Invalid URL input:", err)
			os.Exit(1)
		}
		RSSFFS.Run(pageURL.String(), category, debug, clearCategoryFeeds, conf)
	},
}

// rootCmdPreRun performs setup operations before executing the root command.
//
// This function is called before both the root command and any subcommands
// execute. It configures the logging level based on the debug flag setting.
//
// When debug mode is enabled via the --debug flag, logrus is configured
// to output debug-level messages, providing detailed information about
// the RSS feed discovery and subscription process.
//
// Parameters:
//   - cmd: The cobra command being executed
//   - args: Command-line arguments (unused in this function)
func rootCmdPreRun(cmd *cobra.Command, args []string) {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

// Execute starts the command-line interface execution.
//
// This is the main entry point called from main.go to begin command processing.
// It executes the root command and handles any errors that occur during execution.
//
// If command execution fails, it prints the error message to stdout and
// exits the program with status code 1. This follows standard Unix conventions
// for command-line tool error handling.
//
// Example:
//
//	func main() {
//		cmd.Execute()
//	}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// init initializes the command-line interface during package loading.
//
// This function performs the following setup operations:
//   - Defines persistent flags that are available to all commands
//   - Sets up command-specific flags for the root command
//   - Registers subcommands (man pages and version information)
//
// Persistent flags defined:
//   - debug (-d, --debug): Enables debug-level logging
//   - clearCategoryFeeds (-r, --clearCategoryFeeds): Clears existing feeds before adding new ones
//   - category (-c, --category): Specifies RSS reader category for new feeds
//
// The flags are persistent, meaning they're inherited by all subcommands.
func init() {
	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug-level logging")
	rootCmd.PersistentFlags().BoolVarP(&clearCategoryFeeds, "clearCategoryFeeds", "r", false, "Delete all feeds within category before subscribing to new feeds")
	rootCmd.PersistentFlags().StringVarP(&category, "category", "c", "", "RSS reader category name to assign new feeds to")

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
