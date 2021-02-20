package plugin

// To install External plugins, implement your asuran.go from asurand.go,
// and `import _ "<external/plugin>"`.
// The plugin should be registered in its init().

import _ "github.com/benbearchen/asuran/web/proxy/plugin/chaosjson"
