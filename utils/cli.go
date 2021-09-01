package utils

type RunCmd struct {
	Config string `arg:"-c,--config" default:"config.json" help:"Path to configuration file"`
}

type CreateCmd struct {
	Tag     string `arg:"-t,--tag,required" help:"Tag specifying which default config should be created."`
	OutFile string `arg:"-o,--out" default:"config.json" help:"Output filename for config"`
}

var Args struct {
	Run    *RunCmd    `arg:"subcommand:run"`
	Create *CreateCmd `arg:"subcommand:create"`
}
