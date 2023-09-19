package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/alibaba/pairec/v2/pairecmd/app/options"
	"github.com/alibaba/pairec/v2/pairecmd/commands"
	"github.com/alibaba/pairec/v2/pairecmd/log"
)

func NewProjectCommand(v *viper.Viper, rootcfg *options.RootConfiguration) *cobra.Command {
	cfg := &options.ProjectConfiguration{}
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Create a new pairec application ",
		Aliases: []string{"pr"},
		Run: func(cmd *cobra.Command, args []string) {
			// rootcfg.ParseConfigFromViper(v)
			if err := commands.Project(rootcfg, cfg); err != nil {
				log.Error(err.Error())
			}
		},
	}
	cfg.AddFlags(cmd.Flags())
	return cmd
}

func NewPairecCommand() *cobra.Command {
	cfg := &options.RootConfiguration{}
	rootcmd := &cobra.Command{
		Use:              "pairecmd",
		Long:             `pairecmd is a command line tool  for managing your Pairec Application`,
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.ShowVersion {
				fmt.Println("pairecmd version 1.0.0")
			} else {
				cmd.Usage()
			}
		},
	}
	cfg.AddFlags(rootcmd.Flags())

	v := viper.New()
	rootcmd.AddCommand(NewProjectCommand(v, cfg))
	/**
	rootcmd.AddCommand(NewUpdateCommand(v, cfg))
	rootcmd.AddCommand(NewListCommand(v, cfg))
	rootcmd.AddCommand(NewListAllCommand(v, cfg))
	**/

	return rootcmd
}
