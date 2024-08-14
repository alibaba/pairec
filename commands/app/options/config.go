package options

import "github.com/spf13/pflag"

type RootConfiguration struct {
	ShowVersion bool // Show version information
}

func (cfg *RootConfiguration) AddFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&cfg.ShowVersion, "version", "v", false,
		"show version")
	// flags.AddGoFlagSet(flag.CommandLine)
}

type ProjectConfiguration struct {
	Name string // project name
}

func (cfg *ProjectConfiguration) AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&cfg.Name, "name", "", "",
		"project name")
}
