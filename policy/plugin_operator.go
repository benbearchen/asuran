package policy

type PluginContext struct {
	ProfileIP string // may be empty.
	TargetURL string
	Log       func(statusCode int, postBody, content []byte, err error)
}

type PluginOperator interface {
	Update(context *PluginContext, pluginName string, p *PluginPolicy)
	Remove(context *PluginContext, pluginName string)
	Reset(context *PluginContext, pluginName string)
}
