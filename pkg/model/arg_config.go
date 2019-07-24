package model

// ArgConfig is the settings for passing arguments to the command
type ArgConfig struct {
	ConfigFile string `arg:"-c" help:"Config File to use (default: ~/.aws/route53-ddns.json)"`
	StateFile  string `arg:"-t" help:"State File to use (default: ~/.aws/route53-ddns-state.json)"`
	Setup      bool   `arg:"-s" help:"Setup wizard"`
	Force      bool   `arg:"-f" help:"Force update Route53 even if IP hasn't changed"`
	Quiet      bool   `arg:"-q" help:"Only print errors"`
	Verbose    bool   `arg:"-v" help:"Print a lot more info"`
}

// Description returns a simple description of the command
func (ArgConfig) Description() string {
	return "Update route53 DNS record with your current IP address"
}
