asuran
=========

Asuran is a configurable web proxy with DNS redirection.  [Asuran](http://en.wikipedia.org/wiki/Asuran_%28Stargate%29) is a race in [Stargate Atlantis](http://en.wikipedia.org/wiki/Stargate_Atlantis).

Asuran 是一个使用了 DNS 来实现的 HTTP 透明代理服务，可以配置代理的 URL 及操作。

Asuran 使用 golang 实现，使用 [miekg godns](https://github.com/miekg/dns/) 实现 DNS 服务。


Features 特性
---------

* Profile for each client
* DNS, set a domain to be passed, blocked or redirected to Asuran
* URL, set a url to be dropped, delayed, cached, or <i>overwritten</i>
* History, look url's response content


Build 编译
---------
* Install go 1.2 (can download from [go download page](https://code.google.com/p/go/downloads/list)), add go to PATH
* Set environment GOPATH, witch specifies your go projects dir.  Then mkdir %GOPATH%\src or $GOPATH/src. (see [go code style](http://golang.org/doc/code.html))
* Get sources.  You should have git.  Then:

### \*NIX:
    $ go get github.com/miekg/dns
    $ go get github.com/benbearchen/asuran
### or Windows msys-git:
    $ cd $GOPATH/src
    $ git clone https://github.com/miekg/dns.git github.com/miekg/dns
    $ git clone https://github.com/benbearchen/asuran.git github.com/benbearchen/asuran

* Build & Run:

### \*NIX
    $ cd $GOPATH/src/github.com/benbearchen/asuran
    $ go build asuran.go
    $ ./asuran
### Windows:
    \> cd %GOAPTH%\src\github.com\benbearchen\asuran
    \> go build asuran.go
    \> asuran

* Run asuran anywhere?  Just copy executable asuran\[.exe\] and dir template/.
* Maybe you should open UDP port 53 and TCP port 80 in firewall
* Run and visit asuran host for more information


References 参考
---------

* [golang](http://golang.org/), [Effective Go](http://golang.org/doc/effective_go.html)
* golang [dns](https://github.com/miekg/dns/) by [Miek Gieben](http://www.miek.nl/)
* [《Go Web 编程》](https://github.com/astaxie/build-web-application-with-golang/) by AstaXie （亦是 [beego](http://beego.me/) 作者）
