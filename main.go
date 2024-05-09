package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// 定义目标服务器的地址
	targetURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 创建自定义的处理函数，用于修改请求或响应
	proxy.ModifyResponse = func(response *http.Response) error {
		// 在这里可以对响应进行修改，如添加头部、修改状态码等
		response.Header.Set("X-Custom-Header", "Gateway Server")
		return nil
	}

	// 创建网关服务器
	gatewayServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 在这里可以对请求进行修改，如添加头部、修改路径等
		fmt.Println("Received request:", r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// 启动网关服务器
	fmt.Println("Gateway server is running on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", gatewayServer))
}
