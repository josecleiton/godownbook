package config

type pipeFmt int

const (
	JSON pipeFmt = iota
	YAML
)

type Config struct {
	outDir      string
	outDirBib   string
	defaultRepo string
	pipeCmd     string
	pipeFormat  pipeFmt
	termUi      bool
}
