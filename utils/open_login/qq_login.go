package open_login

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type QQInfo struct {
	Nickname string `json:"nickname"` //昵称
	Gender   string `json:"gender"`   //性别
	Avator   string `json:"avator"`   //头像
	OpenID   string `json:"opne_id"`
}

// QQLogin 用于存储QQ登录的相关信息。
// 它包含了QQ登录过程中所需的各种参数和令牌，用于进行登录验证和授权。
type QQLogin struct {
	appID       string
	appKey      string
	redirect    string
	code        string
	accessToken string
	openID      string
}

// QQConfig 用于配置QQ登录的参数。
// 它包含了应用的ID、密钥和重定向URI，这些是与QQ登录平台交互所必需的配置信息。
type QQConfig struct {
	AppID    string
	AppKey   string
	Redirect string
}

func NewQQLogin(config QQConfig, code string) (qqInfo QQInfo, err error) {
	qqLogin := &QQLogin{
		appID:    config.AppID,
		appKey:   config.AppKey,
		redirect: config.Redirect,
		code:     code,
	}
	err = qqLogin.GetAccessToken()
	if err != nil {
		return qqInfo, err
	}
	err = qqLogin.GetOpenID()
	if err != nil {
		return qqInfo, err
	}
	qqInfo, err = qqLogin.GetUserInfo()
	if err != nil {
		return qqInfo, err
	}
	qqInfo.OpenID = qqLogin.openID
	return qqInfo, nil
}

// GetAccessToken 通过授权码换取QQ登录的访问令牌。
// 该方法实现了QQ登录流程中的重要一步，即使用授权码向QQ接口请求访问令牌。
// 返回错误时，表示获取访问令牌的过程中发生了问题。
// GetAccessToken 用于获取QQ登录的access_token。
func (qq *QQLogin) GetAccessToken() error {
	// 构建请求参数，包括grant_type、client_id、client_secret、code和redirect_uri。
	// 这些参数是向QQ接口请求访问令牌所必需的。
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", qq.appID)
	params.Add("client_secret", qq.appKey)
	params.Add("code", qq.code)
	params.Add("redirect_uri", qq.redirect)

	// 解析QQ授权接口的URL。
	u, err := url.Parse("https://graph.qq.com/oauth2.0/token")
	if err != nil {
		return err
	}

	// 设置URL的查询参数。
	u.RawQuery = params.Encode()

	// 发起HTTP GET请求，获取访问令牌。
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 解析响应体中的查询字符串，获取访问令牌。
	qs, err := parseQS(res.Body)
	if err != nil {
		return err
	}

	// 从解析结果中提取访问令牌，并保存到QQLogin对象中。
	qq.accessToken = qs[`"access_token"`][0]
	return nil
}

// GetOpenID 通过访问QQ OAuth2.0接口获取用户的openid
// 该方法使用已有的access_token来请求QQ接口，解析响应以获取openid，并将其保存在QQLogin实例中。
// 返回值为error类型，用于表示操作过程中可能出现的错误。
// GetOpenID 获取openid
func (qq *QQLogin) GetOpenID() error {
	// 构造请求URL，用于获取用户的openid
	// URL中包含access_token参数，该参数是QQ登录授权过程中获取的临时访问令牌
	u, err := url.Parse(fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", qq.accessToken))
	if err != nil {
		return err
	}

	// 发起HTTP GET请求，获取QQ接口的响应
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}

	// 从响应中提取并解析openid
	// 解析过程由getOpenID函数完成，该函数的细节不在这里说明
	openID, err := getOpenID(res.Body)
	if err != nil {
		return err
	}

	// 将获取到的openid保存在QQLogin实例中，供后续使用
	qq.openID = openID
	return nil
}

// GetUserInfo 通过QQ登录获取用户信息。
// 这个方法使用QQ登录SDK的访问令牌(accessToken)、应用ID(appID)和用户唯一标识(openID)
// 向QQ接口请求用户的详细信息。
// 返回值qqInfo包含从QQ接口获取的用户信息。
// 如果发生错误，错误信息将通过第二个返回值返回。
// GetUserInfo 获取用户信息
func (qq *QQLogin) GetUserInfo() (qqInfo QQInfo, err error) {
	// 构建请求参数
	params := url.Values{}
	params.Add("access_token", qq.accessToken)
	params.Add("oauth_consumer_key", qq.appID)
	params.Add("openid", qq.openID)

	// 解析请求URL
	u, err := url.Parse("https://graph.qq.com/user/get_user_info")
	if err != nil {
		return qqInfo, err
	}

	// 设置请求的查询参数
	u.RawQuery = params.Encode()

	// 发起HTTP GET请求获取用户信息
	res, err := http.Get(u.String())
	if err != nil {
		return qqInfo, err
	}
	defer res.Body.Close()

	// 读取响应体
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return qqInfo, err
	}

	// 解析JSON响应，填充用户信息结构体
	err = json.Unmarshal(data, &qqInfo)
	if err != nil {
		return qqInfo, err
	}

	return qqInfo, nil
}

// parseQS 从给定的 io.Reader 中读取并解析查询字符串，返回一个键值对的映射。
// r: 查询字符串的来源 io.Reader。
// 返回值:
// val: 解析后的键值对映射。
// err: 解析过程中遇到的任何错误。
// parseQS从HTTP响应的正文解析为键值对的形式
func parseQS(r io.Reader) (val map[string][]string, err error) {
	// 使用 readAll 从 r 中读取所有数据，并使用 url.ParseQuery 解析查询字符串。
	val, err = url.ParseQuery(readAll(r))
	if err != nil {
		return val, err
	}
	return val, nil
}

// getOpenID 从HTTP响应体中提取openid
// 参数 r: HTTP响应体的读取器
// 返回值:
//
//	string: 提取到的openid
//	error: 如果未找到openid，则返回错误信息
//
// getOpenID从http响应中获取openID
func getOpenID(r io.Reader) (string, error) {
	// 读取整个HTTP响应体
	body := readAll(r)
	// 查找openid字符串的起始位置
	start := strings.Index(body, `"openid"`) + len(`"openid"`)
	// 如果未找到openid字符串，返回错误
	if start == -1 {
		return "", errors.New("openid not found")
	}
	// 查找openid字符串的结束位置
	end := strings.Index(body[start:], `"`)
	// 如果未找到openid字符串的结束位置，返回错误
	if end == -1 {
		return "", errors.New("openid not found")
	}
	// 从起始位置到结束位置提取openid，并返回
	return body[start : start+end], nil
}

// readAll 从给定的 io.Reader 中读取所有数据，并将其转换为字符串。
// r: 要读取数据的来源 io.Reader。
// 返回值:
// string: 从 io.Reader 中读取并转换的所有数据的字符串表示。
// readAll 读取所有数据并将其转换为字符串
func readAll(r io.Reader) string {
	// 使用 io.ReadAll 从 r 中读取所有数据。
	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	// 将读取的数据转换为字符串并返回。
	return string(b)
}
