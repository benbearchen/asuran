package policy

const pluginKeyword = "plugin"

type PluginPolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(pluginKeyword, "external plugin name", func(name string) (Policy, error) {
		return &PluginPolicy{stringPolicy{pluginKeyword, name, func(name string) string {
			return "通过插件 " + name + " 处理"
		}}}, nil
	}))
}

func (p *PluginPolicy) Name() string {
	return p.Value()
}
