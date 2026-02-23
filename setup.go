package naptrloadbalance

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("naptr_loadbalance", setup) }

func setup(c *caddy.Controller) error {
	c.Next()

	single := false
	if c.NextArg() {
		switch c.Val() {
		case "single":
			single = true
		default:
			return plugin.Error("naptr_loadbalance", c.Errf("unknown argument '%s', expected 'single'", c.Val()))
		}
	}
	if c.NextArg() {
		return plugin.Error("naptr_loadbalance", c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return NaptrLoadBalance{Next: next, Single: single}
	})

	return nil
}
