
Build 编译
---------
### Install go 1.2
install go 1.2 (can download from [go download page](https://code.google.com/p/go/downloads/list)), add go's bin directory to PATH
### Set environment GOPATH
Set environment GOPATH, witch specifies your go projects dir.  Then mkdir %GOPATH%\src or $GOPATH/src. (see [go code style](http://golang.org/doc/code.html))
### Get sources
You should have git.  Then, For \*NIX:

    $ go get github.com/miekg/dns
    $ go get github.com/benbearchen/asuran
For Windows msys-git(if your gotools can't find git):

    $ cd $GOPATH/src
    $ git clone https://github.com/miekg/dns.git github.com/miekg/dns
    $ git clone https://github.com/benbearchen/asuran.git github.com/benbearchen/asuran

### Build & Run:

Build asuran.go in asuran's directory.  Get a executable file asuran\[.exe\].  Run asuran need the ./template directory.
For \*NIX:

    $ cd $GOPATH/src/github.com/benbearchen/asuran
    $ go build asuran.go
    $ ./asuran
For Windows:

    \> cd %GOAPTH%\src\github.com\benbearchen\asuran
    \> go build asuran.go
    \> asuran

Run asuran in other place?  Just copy executable asuran\[.exe\] and dir ./template.
### Firewall
Asuran needs udp port 53(for DNS server), tcp port 80(for asuran HTTP server), and other HTTP ports. Maybe you should open UDP port 53 and TCP port 80 in firewall.
### Runtime usage
Run asuran, then you'll get a host of asuran. Visit it for more informations.


