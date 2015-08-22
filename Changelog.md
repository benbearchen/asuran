Changelog of asuran 
=========

#### asuran 0.2.4 *** 2015-08-23
DNS
    + can disable dns module
    + show global dns history
    + `null` says domain has no ip
HTTP
    + can disable 302, 304
    + can set/remove content-type of response 
    + can timeout the body of response
Profile
    + can delete profile from root console
    + can add/remove operators
    + copy default dns when create
    + can delete a stored content, or modify after created
etc.
    + can speed tcp (hijack from http)


#### asuran 0.2.3 *** 2014-11-10
+ support tcp override http
* fix bug of version code
* fix bug of caching gzip content without ungzip
* fix bug of timeout 0
* update index and usage introduction


#### asuran 0.2.2 *** 2014-10-01
+ speed limit
+ manually bind http port
* fix profile page's proxy button
* fix ranged request return 200 but 206 after 302 redirection


#### asuran 0.2.1 *** 2014-05-14
+ score the url matching, and match the highest command


#### asuran 0.2 *** 2014-04-19
+   URL Proxy Action:  
    *   Matching URL with wildcard  
    *   Delay or drop HTTP connection  
    *   Cache, rewrite or redirect HTTP response  
+   DNS Action:  
    *   Run as DNS server  
    *   Matching domain with wildcard  
    *   Redirect domain's IP  
    *   Block domain response  
+   Profile:  
    *   Profile for each device  
    *   Log action history  
    *   Export command profile  


#### asuran 0.1 *** 2013-12-27
+ create asuran 0.1

