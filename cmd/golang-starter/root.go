package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "go.uber.org/automaxprocs"

	"github.com/toozej/golang-starter/internal/math"
	"github.com/toozej/golang-starter/pkg/config"
	"github.com/toozej/golang-starter/pkg/man"
	"github.com/toozej/golang-starter/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:              "golang-starter",
	Short:            "golang starter examples",
	Long:             `Examples of using math library, cobra and viper modules in golang`,
	PersistentPreRun: rootCmdPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.Config.ConfigVar)

		addMessage := math.Add(1, 2)
		fmt.Println(addMessage)

		subMessage := math.Subtract(2, 2)
		fmt.Println(subMessage)
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
	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")

	// load application configurations
	if err := config.LoadConfig("./config"); err != nil {
		panic(fmt.Errorf("invalid application configuration: %s", err))
	}

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
