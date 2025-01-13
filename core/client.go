package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
}

func NewClient(accessKeyID, accessKeySecret, region string) *Client {
	return &Client{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Region:          region,
	}
}

func (c *Client) DoRequest(method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	// 创建 HTTP 请求
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 添加签名和认证信息
	c.AddSignature(req, headers)

	// 执行请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) AddSignature(req *http.Request, headers map[string]string) {
	// 添加公共头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	// 添加用户自定义头部
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 计算签名
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", req.Method, req.Header.Get("Content-Type"), req.Header.Get("Date"), req.URL.Path, req.URL.RawQuery)
	mac := hmac.New(sha1.New, []byte(c.AccessKeySecret))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 设置 Authorization 头部
	authHeader := fmt.Sprintf("ACS %s:%s", c.AccessKeyID, signature)
	req.Header.Set("Authorization", authHeader)
}
