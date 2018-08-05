package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/solvent-io/zuild/cli"
	"github.com/solvent-io/zuild/zuild"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ZuildCommand struct {
	*cobra.Command
	*cli.Ui

	zuildFlagSet  *pflag.FlagSet
	zfFlagSet     *pflag.FlagSet
	zuildFileInit *zuild.ZuildFileInit
}

func NewZuildCommand() *ZuildCommand {
	cmd := &ZuildCommand{}
	cmd.Command = &cobra.Command{}
	cmd.Ui = cli.NewUi()
	cmd.Use = "zuild"
	cmd.Short = "zuild task tool"
	cmd.Long = "zuild task tool"
	cmd.PreRunE = cmd.setup
	cmd.RunE = cmd.run

	// Flagsets
	cmd.zuildFlagSet = pflag.NewFlagSet("zuild", pflag.ContinueOnError)
	cmd.zfFlagSet = pflag.NewFlagSet("zf", pflag.ContinueOnError)

	// Functional flags
	cmd.zuildFlagSet.Bool("no-color", false, "Disable color")
	cmd.zuildFlagSet.BoolP("Graph", "G", false, "Graph tasks")
	cmd.zuildFlagSet.BoolP("List", "L", false, "List tasks")
	cmd.zuildFlagSet.StringP("Path", "P", "", "Path to Zuildfile")
	cmd.zuildFlagSet.BoolP("Verbose", "V", false, "Verbose output")
	cmd.Flags().AddFlagSet(cmd.zuildFlagSet)

	// Bind custom usage template
	cmd.SetUsageTemplate(cmd.UsageTemplate())
	cobra.AddTemplateFunc("HasZuildHelp", cmd.hasHelp)
	cobra.AddTemplateFunc("ZuildHelpTitle", cmd.helpTitle)
	cobra.AddTemplateFunc("ZuildHelpContent", cmd.helpContent)
	cobra.AddTemplateFunc("HasZuildFlags", cmd.hasFlags)
	cobra.AddTemplateFunc("ZuildFlags", cmd.flags)

	// Init help early
	cmd.InitDefaultHelpFlag()
	cmd.ParseFlags(os.Args[1:])

	path, _ := cmd.Flags().GetString("Path")

	zfPath, err := cmd.zfPath(path)
	if err != nil {
		cmd.Error(fmt.Sprint("Error: ", err.Error()))
		cmd.Usage()
		os.Exit(1)
	}

	cmd.zuildFileInit, err = zuild.ParseZuildFile(zfPath)
	if err != nil {
		cmd.Fatal(fmt.Sprint("Error: ", err.Error()))
	}

	for _, arg := range cmd.zuildFileInit.Args {
		cmd.zfFlagSet.StringP(arg.Name, arg.Short, "", arg.Usage)
	}

	cmd.Flags().AddFlagSet(cmd.zfFlagSet)

	return cmd
}

func (z *ZuildCommand) zfPath(zfPath string) (string, error) {
	wdr, err := os.Getwd()
	if err != nil {
		return "", err
	}

	wd, err := filepath.Abs(wdr)
	if err != nil {
		return "", err
	}

	if zfPath == "" {
		zfPath = path.Join(wd, zuild.DefaultZfPath)
	} else {
		zfPath, err = filepath.Abs(zfPath)
		if err != nil {
			return "", err
		}
	}

	// If flag path is a directory append DefaultZfPath
	stat, err := os.Stat(zfPath)
	if err != nil {
		return "", errors.New(fmt.Sprint("no Zuildfile found at: ", zfPath))
	}
	if stat.IsDir() {
		zfPath = path.Join(zfPath, zuild.DefaultZfPath)
	}

	if _, err := os.Stat(zfPath); os.IsNotExist(err) {
		return "", errors.New(fmt.Sprint("no Zuildfile found at: ", zfPath))
	}

	return zfPath, nil
}

func (z *ZuildCommand) setup(cmd *cobra.Command, args []string) error {
	color, err := cmd.Flags().GetBool("no-color")

	z.NoColor(color)

	return err
}

func (z *ZuildCommand) run(cmd *cobra.Command, args []string) error {
	zuild, err := zuild.New(cmd, z.zuildFileInit)
	if err != nil {
		return err
	}

	// Bind event handlers
	zuild.On("out", func(message string) {
		z.Ui.Out(message)
	})

	zuild.On("task.header", func(message string) {
		z.Ui.Info(fmt.Sprint("* ", message))
	})

	zuild.On("action.header", func(message string) {
		z.Ui.Info(fmt.Sprint("  ", message))
	})

	zuild.On("action.error", func(message string) {
		z.Ui.Error(fmt.Sprint("  x", message))
	})

	zuild.On("action.warn", func(message string) {
		z.Ui.Error(fmt.Sprint("  ~", message))
	})

	zuild.On("action.verbose.header", func(message string) {
		z.Ui.Out(fmt.Sprint("  > ", message))
	})

	zuild.On("action.verbose.content", func(message string) {
		z.Ui.Out(fmt.Sprint("    ", message))
	})

	list, _ := cmd.Flags().GetBool("List")

	if list {
		return zuild.List()
	}

	graph, _ := cmd.Flags().GetBool("Graph")

	if graph {
		return zuild.Graph(cmd.Flags().Arg(0))
	}

	return zuild.Run(cmd.Flags().Arg(0))
}

func (z *ZuildCommand) helpContent() string {
	if z.zuildFileInit != nil {
		return z.zuildFileInit.Help.Content
	}

	return ""
}

func (z *ZuildCommand) helpTitle() string {
	if z.zuildFileInit != nil {
		return z.zuildFileInit.Help.Title
	}

	return ""
}

func (z *ZuildCommand) hasHelp() bool {
	if z.helpContent() != "" {
		return true
	}

	return false
}

func (z *ZuildCommand) hasFlags() bool {
	if z.zuildFileInit != nil {
		if len(z.zuildFileInit.Args) > 0 {
			return true
		}
	}

	return false
}

func (z *ZuildCommand) flags(class string) string {
	if class == "local" {
		return z.zuildFlagSet.FlagUsages()
	}

	if class == "file" {
		return z.zfFlagSet.FlagUsages()
	}

	return ""
}

func (z *ZuildCommand) UsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{ZuildFlags "local" | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}{{ if HasZuildHelp }}

{{ZuildHelpTitle}}

{{ZuildHelpContent}}{{end}}{{ if HasZuildFlags }}
Flags:
{{ZuildFlags "file" | trimTrailingWhitespaces}}{{end}}
`
}
