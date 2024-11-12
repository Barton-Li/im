package main

import (
	"bytes"
	"encoding/json"
	"fim/common/etcd"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Date struct {
	Code int    `json:"code"`
	Date any    `json:"data"`
	Msg  string `json:"msg"`
}

func FilResponse(msg string, res http.ResponseWriter) {
	response := Date{
		Code: 7,
		Msg:  msg,
	}
	byteData, _ := json.Marshal(response)
	_, err := res.Write(byteData)
	if err != nil {
		return
	}

}

// 验证
// auth 对请求进行认证，通过向authAddr发送请求来验证当前请求的合法性。
// 参数authAddr是认证服务的地址。
// 参数res和req分别是HTTP响应和请求对象。
// 返回值ok表示认证是否成功。
func auth(authAddr string, res http.ResponseWriter, req *http.Request) (ok bool) {
	// 创建一个新的HTTP请求，方法为POST，目标地址为authAddr。
	authReq, _ := http.NewRequest("POST", authAddr, nil)
	// 将当前请求的Header复制到authReq中。
	authReq.Header = req.Header
	// 设置ValidPath头，携带当前请求的URL路径，作为认证的一部分。
	authReq.Header.Set("ValidPath", req.URL.Path)
	// 使用默认的HTTP客户端发送authReq，并获取响应。
	authRes, err := http.DefaultClient.Do(authReq)
	if err != nil {
		// 如果发送请求过程中出现错误，记录错误并返回认证失败。
		logx.Error(err)
		FilResponse("验证失败", res)
		return false
	}
	// 定义响应结构，用于解析authRes的主体。
	type Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			UserID int `json:"userId"`
			Role   int `json:"role"`
		} `json:"date"`
	}

	var authResponse Response
	// 读取authRes的主体。
	byteData, _ := io.ReadAll(authRes.Body)
	// 解析主体为Response结构。
	authErr := json.Unmarshal(byteData, &authResponse)
	if authErr != nil {
		// 如果解析过程中出现错误，记录错误并返回认证失败。
		logx.Error(authErr)
		FilResponse("验证失败", res)
		return
	}
	// 如果响应码不为0，表示认证失败，将响应主体写回res并返回。
	if authResponse.Code != 0 {
		_, err2 := res.Write(byteData)
		if err2 != nil {
			return false
		}
		return
	}
	// 如果认证成功，将用户的ID和角色设置到req的Header中。
	if authResponse.Data != nil {
		req.Header.Set("UserID", fmt.Sprintf("%d", authResponse.Data.UserID))
		req.Header.Set("Role", fmt.Sprintf("%d", authResponse.Data.Role))
	}
	// 返回认证成功。
	return true
}

// proxy 通过指定的代理地址转发请求。
// 参数:
//
//	proxyAddr - 代理服务器的地址。
//	res - 用于向客户端发送响应的http.ResponseWriter。
//	req - 从客户端接收的http.Request。
func proxy(proxyAddr string, res http.ResponseWriter, req *http.Request) {
	// 完整读取客户端的请求体
	// 读取客户端请求体
	byteData, _ := io.ReadAll(req.Body)
	// 创建一个新的请求，用于向代理服务器发送请求
	// 创建新的请求对象，用于代理服务发送请求
	proxyReq, err := http.NewRequest(req.Method, proxyAddr, bytes.NewBuffer(byteData))
	if err != nil {
		// 记录错误并返回错误响应
		logx.Error(err)
		FilResponse("代理失败", res)
		return
	}
	// 复制原请求的头部信息
	proxyReq.Header = req.Header
	// 删除可能存在的ValidPath头部，以避免认证问题
	proxyReq.Header.Del("ValidPath")
	// 使用默认的HTTP客户端发送请求
	response, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		// 记录错误并返回错误响应
		logx.Error(err)
		FilResponse("服务异常", res)
		return
	}
	// 将代理服务器的响应拷贝到客户端
	io.Copy(res, response.Body)
}

// gateway 作为请求的入口，负责根据URL路径将请求代理到相应的服务。
// 它首先通过正则表达式解析URL路径，找到对应的服务名，然后从服务映射表中查找服务的地址。
// 如果服务存在，它将构建一个新的HTTP请求并设置相关的请求头，包括转发地址。
// 最后，它将代理请求到目标服务，并将响应返回给原始请求者。
func gateway(res http.ResponseWriter, req *http.Request) {
	regex, _ := regexp.Compile(`/api/(.*?)/`)
	addrList := regex.FindStringSubmatch(req.URL.Path)
	if len(addrList) != 2 {
		res.Write([]byte("err"))
		return
	}
	service := addrList[1]
	addr := etcd.GetAddress(config.Etcd, service+"_api")
	if addr == "" {
		logx.Errorf("%s 不匹配服务", service)
		FilResponse("err", res)
		return
	}
	remoteAddr := strings.Split(req.RemoteAddr, ":")
	authAddr := etcd.GetAddress(config.Etcd, "auth_api")
	authUrl := fmt.Sprintf("http://%s/api/auth/authentication", authAddr)
	proxyUrl := fmt.Sprintf("http://%s%s", addr, req.URL.String())

	logx.Infof("%s %S", remoteAddr[0], proxyUrl)
	if !auth(authUrl, res, req) {
		return
	}
	proxy(proxyUrl, res, req)

}

type Config struct {
	Addr string
	Etcd string
	Log  logx.LogConf
}

var config Config
var configFile = flag.String("f", "settings.yaml", "the config file")

// main函数是程序的入口点
func main() {
	// 先处理命令行参数，确保参数已经被解析
	flag.Parsed()
	// 加载配置文件，如果加载失败，则会panic
	conf.MustLoad(*configFile, &config)
	// 根据配置文件设置日志系统
	logx.SetUp(config.Log)
	// 注册HTTP服务的处理函数
	http.HandleFunc("/", gateway)
	// 输出启动信息，包括服务地址
	fmt.Println("Starting FIM Gateway...", config.Addr)
	// 监听指定地址，启动HTTP服务
	http.ListenAndServe(config.Addr, nil)
}
