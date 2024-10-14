package cmd

import (
	"fmt"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/toozej/RSSFFS/internal/RSSFFS"
	"github.com/toozej/RSSFFS/pkg/man"
	"github.com/toozej/RSSFFS/pkg/version"
)

var category string

var rootCmd = &cobra.Command{
	Use:              "RSSFFS [pageURL]",
	Short:            "RSS Feed Finder [and] Subscriber",
	Long:             `Automatically find and subscribe to RSS feeds found on inputted URL, and on URLs mentioned on the inputted URL.`,
	Args:             cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	PersistentPreRun: rootCmdPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		inputURL := args[0]
		pageURL, err := url.ParseRequestURI(inputURL)
		if err != nil {
			fmt.Println("Invalid URL input:", err)
			os.Exit(1)
		}
		RSSFFS.Run(pageURL.String(), category, viper.GetBool("debug"), viper.GetBool("clearCategoryFeeds"))
	},
}

func rootCmdPreRun(cmd *cobra.Command, args []string) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return
	}
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	_, err := maxprocs.Set()
	if err != nil {
		log.Error("Error setting maxprocs: ", err)
	}

	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")
	rootCmd.PersistentFlags().BoolP("clearCategoryFeeds", "r", false, "Delete all feeds within category before subscribing to new feeds")
	rootCmd.PersistentFlags().StringVarP(&category, "category", "c", "", "RSS reader category name to assign new feeds to")

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
