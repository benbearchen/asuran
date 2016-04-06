asuran
=========

Asuran is a configurable web proxy with DNS redirection.  [Asuran](http://en.wikipedia.org/wiki/Asuran_%28Stargate%29) is a race in [Stargate Atlantis](http://en.wikipedia.org/wiki/Stargate_Atlantis).

Asuran 是一个使用了 DNS 来实现的 HTTP 透明代理服务，可以配置代理的 URL 及操作。当然用做标准 HTTP 代理也是可以的。

Asuran 使用 golang 实现，使用 [miekg godns](https://github.com/miekg/dns/) 实现 DNS 服务。


Features 特性
---------

* DNS Server
    * run like a real DNS Server
    * pass, redirect(like /etc/hosts), `block`, or `null`(can's found a IP) a request
    * redirect to asuran's HTTP Proxy(like a transparent HTTP proxy)
    * `delay` a response in const or rand durations
* HTTP Proxy
    * Standard or Transparent HTTP Proxy
    * proxy and modify HTTP's content:
        * set `status` code
        * set/remove/empty `content-type`
        * `redirect` to a new URL
        * `cache` URL's content and return in the future
        * `map` the content(including the HTTP headers) from another URL
        * `restore`|`rewrite` the content as response body
        * hijack then response as TCP without HTTP format
        * force the `chunked` to be enabled or disabled
        * force the connected `host`
        * force the 304 to be disabled
        * make choice of executing 302 in server or client
    * change the speed or rtt of HTTP:
        * `drop` the response in some duration
        * waiting a duration before HTTP headers then `timeout` or only `delay`
        * waiting another duration before HTTP body(after HTTP headers, also)
        * limit the speed of sending HTTP body
        * timeout in a duration, from beginning of sending body
* Profile for each device
    * each device has its own profile
    * config policy commands of domains and URLs
    * request histories
    * add/remove operators
* Capture websocket(without modify)


References 参考
---------

* [golang](http://golang.org/), [Effective Go](http://golang.org/doc/effective_go.html)
* golang [dns](https://github.com/miekg/dns/) by [Miek Gieben](http://www.miek.nl/)
* [《Go Web 编程》](https://github.com/astaxie/build-web-application-with-golang/) by AstaXie （亦是 [beego](http://beego.me/) 作者）


Thanks
========

Thanks fatcowfeng for helping asuran.  Bless.
