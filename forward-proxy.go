package main

/**
正向代理
 */

import (
	"net/http"
	"fmt"
	"net"
	"strings"
	"io"
)

type Proxy struct {
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request){
	fmt.Printf("接受请求 %s %s %s\n",req.Method,req.Host,req.RemoteAddr)

	transport:=http.DefaultTransport

	// 第一步： 代理接受到客户端的请求，复制原来的请求对象，并根据数据配置新请求的各种参数(添加上X-Forward-For头部等)
	outReq:=new(http.Request)
	*outReq = *req // 这只是一个浅层拷贝

	clientIP,_,err:=net.SplitHostPort(req.RemoteAddr)
	if err==nil{
		prior,ok:=outReq.Header["X-Forwarded-For"]
		if ok {
			clientIP = strings.Join(prior,", ")+", "+clientIP
		}
		outReq.Header.Set("X-Forwarded-For",clientIP)
	}

	// 第二步： 把新请求复制到服务器端，并接收到服务器端返回的响应
	res,err:=transport.RoundTrip(outReq)
	if err!=nil{
		w.WriteHeader(http.StatusBadGateway) // 502
		return
	}

	// 第三步：代理服务器对响应做一些处理，然后返回给客户端
	for key,value:=range res.Header{
		for _,v:=range value{
			w.Header().Add(key,v)
		}
	}

	w.WriteHeader(res.StatusCode)
	io.Copy(w,res.Body)
	res.Body.Close()
}

func main() {
	fmt.Println("Serve on :8080")
	http.Handle("/",&Proxy{})
	http.ListenAndServe("0.0.0.0:8080",nil)
}

// 代码运行之后，会在本地的 8080 端口启动代理服务。修改浏览器的代理为 127.0.0.1：:8080
// 再访问网站，可以验证代理正常工作，也能看到它在终端打印出所有的请求信息。
