package agentskill

import "embed"

//go:embed assets/sonacli/SKILL.md assets/sonacli/agents/openai.yaml
var embeddedAssets embed.FS

type embeddedAsset struct {
	source  string
	path    string
	targets []Target // empty means all targets
}

var embeddedSkillAssets = []embeddedAsset{
	{
		source: "assets/sonacli/SKILL.md",
		path:   "SKILL.md",
	},
	{
		source:  "assets/sonacli/agents/openai.yaml",
		path:    "agents/openai.yaml",
		targets: []Target{TargetCodex},
	},
}
