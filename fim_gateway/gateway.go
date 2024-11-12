package main

import (
	"encoding/json"
	"fim/common/etcd"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

type BaseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Date any    `json:"data"`
}

func FilResponse(msg string, res http.ResponseWriter) {
	response := BaseResponse{Code: 7, Msg: msg}
	byteData, _ := json.Marshal(response)
	res.Write(byteData)
}

// auth 对请求进行认证，通过向认证服务发送请求来验证token的有效性。
// 如果认证成功，将在原始请求头中添加用户ID和角色信息。
// 参数:
//
//	authAddr - 认证服务的地址。
//	res - 用于向客户端发送响应的http.ResponseWriter。
//	req - 从客户端接收的http.Request。
//
// 返回值:
//
//	ok - 认证是否成功的布尔值。
func auth(authAddr string, res http.ResponseWriter, req *http.Request) (ok bool) {
	// 创建一个新的HTTP请求来向认证服务发送认证请求。
	authReq, _ := http.NewRequest("POST", authAddr, nil)
	// 将原始请求的头信息复制到认证请求中。
	authReq.Header = req.Header
	// 从URL查询参数中获取token，并设置到认证请求的头信息中。
	token := req.URL.Query().Get("token")
	if token != "" {
		authReq.Header.Set("Token", token)
	}
	// 设置请求的路径到认证请求的头信息中，用于认证服务验证请求的合法性。
	authReq.Header.Set("ValidPath", req.URL.Path)
	// 发送认证请求并处理可能的错误。
	authRes, err := http.DefaultClient.Do(authReq)
	if err != nil {
		logx.Error(err)
		FilResponse("认证服务错误", res)
		return
	}
	// 定义用于解析认证服务响应的结构体。
	// 定义用于解析认证服务响应的结构体。
	type Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			UserID uint `json:"userID"`
			Role   int  `json:"role"`
		} `json:"data"`
	}

	// 解析认证服务的响应。
	var authResponse Response
	byteData, _ := io.ReadAll(authRes.Body)
	authErr := json.Unmarshal(byteData, &authResponse)
	if authErr != nil {
		logx.Error(authErr)
		FilResponse("认证服务响应解析错误", res)
		return
	}

	// 如果认证响应的代码不为0，表示认证失败，将认证服务的响应直接返回给客户端。
	// 认证不通过
	if authResponse.Code != 0 {
		res.Write(byteData)
		return
	}

	// 如果认证成功，将用户ID和角色信息添加到请求的头信息中。
	if authResponse.Data != nil {
		req.Header.Set("User-ID", fmt.Sprintf("%d", authResponse.Data.UserID))
		req.Header.Set("Role", fmt.Sprintf("%d", authResponse.Data.Role))
	}
	// 认证成功，返回true。
	return true
}

var configFile = flag.String("f", "settings.yaml", "the config file")

type Config struct {
	Addr string
	Etcd string
	Log  logx.LogConf
}

var config Config

type Proxy struct {
}

// ServeHTTP 实现了 http.Handler 接口，用于处理所有通过代理的 HTTP 请求。
// 它会根据请求的 URL 路径匹配对应的服务地址，并将请求代理到对应的服务上。
func (Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// 编译正则表达式，用于匹配 URL 中的服务名。
	regex, _ := regexp.Compile(`/api/(.*?)/`)
	// 输出请求的 URL 路径，用于调试。
	fmt.Println(req.URL.Path)
	// 使用正则表达式匹配 URL，获取服务名。
	addrList := regex.FindStringSubmatch(req.URL.Path)
	// 如果匹配结果长度不为2，说明 URL 格式不正确，返回错误响应。
	if len(addrList) != 2 {
		res.Write([]byte("err"))
		return
	}
	// 从匹配结果中提取服务名。
	service := addrList[1]
	// 通过服务名从 etcd 获取服务的地址。
	addr := etcd.GetAddress(config.Etcd, service+"_api")
	// 如果获取不到地址，说明该服务不存在，返回错误响应。
	if addr == "" {
		logx.Errorf("%s 不匹配服务", service)
		FilResponse("err", res)
		return
	}
	// 从请求中获取客户端的地址。
	remoteAddr := strings.Split(req.RemoteAddr, ":")
	// 从 etcd 获取认证服务的地址。
	authAddr := etcd.GetAddress(config.Etcd, "auth_api")
	// 组装认证服务的 URL。
	authUr1 := fmt.Sprintf("http://%s/api/auth/authentication", authAddr)
	// 组装要代理的服务的 URL。
	proxyUr1 := fmt.Sprintf("http://%s%s", addr, req.URL.String())

	// 输出客户端地址和要代理的 URL，用于调试。
	logx.Infof("%s %s", remoteAddr[0], proxyUr1)
	// 调用认证函数，如果认证失败，返回错误响应。
	if !auth(authUr1, res, req) {
		return
	}
	// 解析要代理的服务的地址。
	remote, _ := url.Parse(fmt.Sprintf("http://%s", addr))
	// 创建反向代理对象。
	reverseProxy := httputil.NewSingleHostReverseProxy(remote)
	// 通过反向代理处理请求。
	reverseProxy.ServeHTTP(res, req)
}

// main函数是程序的入口点
func main() {
	// 解析命令行参数
	flag.Parse()
	// 加载配置文件
	conf.MustLoad(*configFile, &config)
	// 设置日志配置
	logx.SetUp(config.Log)
	// 输出服务启动信息
	fmt.Printf("gateway running %s\n", config.Addr)

	// 初始化代理服务
	proxy := Proxy{}
	// 启动HTTP服务监听
	http.ListenAndServe(config.Addr, proxy)
}
