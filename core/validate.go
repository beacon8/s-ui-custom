package core

import (
	"github.com/sagernet/sing-box/option"
)

// ValidateConfig decodes a complete sing-box configuration with the same
// protocol registries the running core uses, without starting listeners or
// dialing anything. It is the preflight the panel needs before it commits a
// batch: a config that fails here would leave the previous core replaced by a
// dead one on restart.
func (c *Core) ValidateConfig(config []byte) error {
	var options option.Options
	return options.UnmarshalJSONContext(c.ctx, config)
}

// ValidateOutbound decodes one outbound with the same registry. It catches
// unknown types, misspelled option names and invalid transports before a batch
// is written.
func (c *Core) ValidateOutbound(config []byte) error {
	var outbound option.Outbound
	return outbound.UnmarshalJSONContext(c.ctx, config)
}
