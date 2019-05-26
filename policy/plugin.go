package policy

import (
	"github.com/benbearchen/asuran/util/cmd"

	"fmt"
	"strings"
)

const pluginKeyword = "plugin"

const (
	settingSubKeyword = "setting"
	setSubKeyword     = "set"
	deleteSubKeyword  = "delete"
)

type PluginSetter struct {
	Key string
	Value string
}

type PluginPolicy struct {
	setting string
	setter  *PluginSetter
	deleter *string
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
	var setter *PluginSetter
	var deleter *string
	var err error
	switch args[0] {
	case settingSubKeyword:
		if len(args) < 3 {
			return nil, args, fmt.Errorf("plugin setting need a setting-string, then a plugin name")
		}

		setting = args[1]
		args = args[2:]
	case setSubKeyword:
		if len(args) < 3 {
			return nil, args, fmt.Errorf("plugin set need <key>=<value>, then a plugin name")
		}

		setter, err = parseKeyValueSet(args[1])
		if err != nil {
			return nil, args, err
		}

		args = args[2:]
	case deleteSubKeyword:
		if len(args) < 3 {
			return nil, args, fmt.Errorf("plugin delete need a <key>, then a plugin name")
		}

		deleter = &args[1]
		args = args[2:]
	default:
	}

	name := args[0]
	args = args[1:]

	return &PluginPolicy{setting, setter, deleter, name}, args, nil
}

func (p *PluginPolicy) Keyword() string {
	return pluginKeyword
}

func (p *PluginPolicy) Command() string {
	cmds := make([]string, 0, 4)
	cmds = append(cmds, pluginKeyword)
	if len(p.setting) > 0 {
		cmds = append(cmds, settingSubKeyword, cmd.Quote(p.setting))
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

func (p *PluginPolicy) Setter() *PluginSetter {
	return p.setter
}

func (p *PluginPolicy) DeleteKey() string {
	if p.deleter != nil {
		return *p.deleter
	} else {
		return ""
	}
}

func (p *PluginPolicy) Feedback(setting string) {
	p.setting = setting
}

func parseKeyValueSet(keyValue string) (*PluginSetter, error) {
	p := strings.IndexByte(keyValue, '=')
	if p == -1 {
		return nil, fmt.Errorf("<key>=<value> `%s' need a `='", keyValue)
	} else if p == 0 {
		return nil, fmt.Errorf("<key>=<value> `%s' need a key", keyValue)
	} else {
		return &PluginSetter{keyValue[:p], keyValue[p+1:]}, nil
	}
}

