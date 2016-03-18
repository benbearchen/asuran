package policy

import (
	"fmt"
	"strings"
)

const pluginKeyword = "plugin"

const settingSubKeyword = "setting"

type PluginPolicy struct {
	setting string
	name    string
}

type pluginPolicyFactory struct {
}

func init() {
	regFactory(new(pluginPolicyFactory))
}

func (f *pluginPolicyFactory) Keyword() string {
	return pluginKeyword
}

func (f *pluginPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) <= 0 {
		return nil, args, fmt.Errorf("need atleast a external plugin name")
	}

	setting := ""
	if args[0] == settingSubKeyword {
		if len(args) < 3 {
			return nil, args, fmt.Errorf("plugin setting need a setting-string, then a plugin name")
		}

		setting = args[1]
		args = args[2:]
	}

	name := args[0]
	args = args[1:]

	return &PluginPolicy{setting, name}, args, nil
}

func (p *PluginPolicy) Keyword() string {
	return pluginKeyword
}

func (p *PluginPolicy) Command() string {
	cmds := make([]string, 0, 4)
	cmds = append(cmds, pluginKeyword)
	if len(p.setting) > 0 {
		cmds = append(cmds, settingSubKeyword, p.setting)
	}

	cmds = append(cmds, p.name)
	return strings.Join(cmds, " ")
}

func (p *PluginPolicy) Comment() string {
	return "通过插件 " + p.name + " 处理"
}

func (p *PluginPolicy) Update(n Policy) error {
	switch n := n.(type) {
	case *PluginPolicy:
		p.setting = n.setting
		p.name = n.name
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (p *PluginPolicy) Name() string {
	return p.name
}

func (p *PluginPolicy) Setting() string {
	return p.setting
}
