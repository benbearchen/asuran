asuran
=========

Asuran is a configurable web proxy with DNS redirection.  [Asuran](http://en.wikipedia.org/wiki/Asuran_%28Stargate%29) is a race in [Stargate Atlantis](http://en.wikipedia.org/wiki/Stargate_Atlantis).

Asuran 是一个使用了 DNS 来实现的 HTTP 透明代理服务，可以配置代理的 URL 及操作。

Asuran 使用 golang 实现，使用 [miekg godns](https://github.com/miekg/dns/) 实现 DNS 服务。


Features 特性
---------

* Profile for each client
* DNS, set a domain to be passed, blocked or redirected to Asuran
* URL, set a url to be dropped, delayed, cached, or overwritten
* History


References 参考
---------

* [golang](http://golang.org/), [Effective Go](http://golang.org/doc/effective_go.html)
* [godns](https://github.com/miekg/dns/) by [Miek Gieben](http://www.miek.nl/)
* [《Go Web 编程》](https://github.com/astaxie/build-web-application-with-golang/) by AstaXie （亦是 [beego](http://beego.me/) 作者）
