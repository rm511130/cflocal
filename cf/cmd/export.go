package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/sclevine/cflocal/local"
)

type Export struct {
	UI     UI
	Stager Stager
	Runner Runner
	FS     FS
	Help   Help
	Config Config
}

type exportOptions struct {
	name, reference string
}

func (e *Export) Match(args []string) bool {
	return len(args) > 0 && args[0] == "export"
}

func (e *Export) Run(args []string) error {
	options, err := e.options(args)
	if err != nil {
		if err := e.Help.Show(); err != nil {
			e.UI.Error(err)
		}
		return err
	}
	droplet, dropletSize, err := e.FS.ReadFile(fmt.Sprintf("./%s.droplet", options.name))
	if err != nil {
		return err
	}
	defer droplet.Close()
	launcher, launcherSize, err := e.Stager.Launcher()
	if err != nil {
		return err
	}
	defer launcher.Close()
	localYML, err := e.Config.Load()
	if err != nil {
		return err
	}
	id, err := e.Runner.Export(&local.RunConfig{
		Droplet:      droplet,
		DropletSize:  dropletSize,
		Launcher:     launcher,
		LauncherSize: launcherSize,
		AppConfig:    getAppConfig(options.name, localYML),
	}, options.reference)
	if err != nil {
		return err
	}
	if options.reference != "" {
		e.UI.Output("Exported %s as %s with ID: %s", options.name, options.reference, id)
	} else {
		e.UI.Output("Exported %s with ID: %s", options.name, id)
	}
	return nil
}

func (*Export) options(args []string) (*exportOptions, error) {
	set := &flag.FlagSet{}
	options := &exportOptions{}
	set.StringVar(&options.reference, "r", "", "")
	if err := set.Parse(args[1:]); err != nil {
		return nil, err
	}
	if set.NArg() != 1 {
		return nil, errors.New("invalid arguments")
	}
	options.name = set.Arg(0)
	return options, nil
}