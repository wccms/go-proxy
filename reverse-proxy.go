package main

import (
	"net/url"
	"net/http/httputil"
	"net/http"
	"math/rand"
	"log"
)

/**
反向代理：编写反向代理按照上面的思路当然没有问题，只需要在第二步的时候，根据之前的配置修改 outReq 的 URL Host 地址可以了。
不过 Golang 已经给我们提供了编写代理的框架： httputil.ReverseProxy 。我们可以用非常简短的代码来实现自己的代理，而且内
部的细节问题都已经被很好地处理了。

实现一个简单的反向代理，它能够对请求实现负载均衡，随机地把请求发送给某些配置好的后端服务器。使用 httputil.ReverseProxy
编写反向代理最重要的就是实现自己的 Director 对象，这是 GoDoc 对它的介绍
Director必须是一个功能，它将请求修改为使用Transport发送的新请求。然后将其响应未经修改地复制回原始客户端。 Director返回
后不得访问提供的请求。

简单翻译的话， Director 是一个函数，它接受一个请求作为参数，然后对其进行修改。修改后的请求会实际发送给服务器端，因此我们编
写自己的 Director 函数，每次把请求的
Scheme 和 Host 修改成某个后端服务器的地址，就能实现负载均衡的效果（其实上面的正向代理也可以通过相同的方法实现）
 */

func NewMultipleHostsReverseProxy(targets []*url.URL) *httputil.ReverseProxy{
 	director:= func(req *http.Request) {
 		target:=targets[rand.Int()*len(targets)]
 		req.URL.Scheme = target.Scheme
 		req.URL.Host = target.Host
 		req.URL.Path = target.Path
	}
 	return &httputil.ReverseProxy{
 		Director:director,
	}
}

func main() {
	proxy:=NewMultipleHostsReverseProxy([]*url.URL{
		{
			Scheme:"http",
			Host:"localhost:9091",
		},
		{
			Scheme:"http",
			Host:"localhost:9092",
		},
	})
	log.Fatal(http.ListenAndServe(":9090",proxy))
}

// 让代理监听在 9090 端口，在后端启动两个返回不同响应的服务器分别监听
// 在 9091 和 9092 端口，通过 curl 访问，可以看到多次请求会返回不同的结果。
// curl http://127.0.0.1:9090
// curl http://127.0.0.1:9090