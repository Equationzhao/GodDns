package Net

import (
	"net/url"

	"GodDns/Util"
)

type proxy = string

type Proxies []proxy

var GlobalProxies = &Proxies{}

func IsProxyValid(proxy proxy) bool {
	_, err := url.Parse(proxy)
	return err == nil
}

func AddProxy(target *Proxies, proxy ...proxy) {
	*target = append(*target, proxy...)
}

func AddProxy2Top(target *Proxies, proxy ...proxy) {
	*target = append(proxy, *target...)
}

func (p *Proxies) GetProxyIter() *Util.Iter[proxy] {
	return Util.NewIter((*[]proxy)(p))
}
