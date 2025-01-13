package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/godrealms/ali-fusion-go-sdk/core"
	"io"
	"net/http"
	"os"
	"time"
)

type OSSClient struct {
	client     *core.Client
	BucketName string
	Endpoint   string
}

func NewOSSClient(client *core.Client, bucketName, endpoint string) *OSSClient {
	return &OSSClient{
		client:     client,
		BucketName: bucketName,
		Endpoint:   endpoint,
	}
}

// UploadFile 上传文件并返回文件访问地址
func (o *OSSClient) UploadFile(objectKey, filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// 获取文件大小
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	// 拼接上传 URL
	url := fmt.Sprintf("https://%s.%s/%s", o.BucketName, o.Endpoint, objectKey)

	// 创建 HTTP 请求
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = fileSize

	// 添加签名
	o.client.AddSignature(req, nil)

	// 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 检查返回状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 返回文件的访问地址
	fileURL := fmt.Sprintf("https://%s.%s/%s", o.BucketName, o.Endpoint, objectKey)
	return fileURL, nil
}

// DownloadFile 下载文件
func (o *OSSClient) DownloadFile(objectKey, destPath string) error {
	url := fmt.Sprintf("https://%s.%s/%s", o.BucketName, o.Endpoint, objectKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 添加签名
	o.client.AddSignature(req, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// ListObjects 列出文件
func (o *OSSClient) ListObjects(prefix string) ([]Object, error) {
	url := fmt.Sprintf("https://%s.%s/?prefix=%s", o.BucketName, o.Endpoint, prefix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 添加签名
	o.client.AddSignature(req, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list objects failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	body, _ := io.ReadAll(resp.Body)
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Contents, nil
}

// DeleteObject 删除文件
func (o *OSSClient) DeleteObject(objectKey string) error {
	url := fmt.Sprintf("https://%s.%s/%s", o.BucketName, o.Endpoint, objectKey)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 添加签名
	o.client.AddSignature(req, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UploadPolicy 临时签名配置
type UploadPolicy struct {
	AccessKeyId string `json:"accessKeyId"`
	Policy      string `json:"policy"`
	Signature   string `json:"signature"`
	Bucket      string `json:"bucket"`
	Endpoint    string `json:"endpoint"`
	Expire      int64  `json:"expire"`
}

// GenerateUploadSignature 生成上传签名
func (o *OSSClient) GenerateUploadSignature(expireSeconds int64, maxFileSize int64, dirPrefix string) (*UploadPolicy, error) {
	// 生成过期时间戳
	expireTime := time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix()

	// 定义 Policy 策略
	policy := map[string]interface{}{
		"expiration": time.Unix(expireTime, 0).UTC().Format(time.RFC3339), // 策略过期时间
		"conditions": []interface{}{
			// 限制文件大小
			[]interface{}{"content-length-range", 0, maxFileSize},
			// 限制文件上传路径前缀
			map[string]string{"key": dirPrefix},
		},
	}

	// 将策略转换为 JSON 并进行 Base64 编码
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy: %v", err)
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyBytes)

	// 使用 AccessKeySecret 对 Policy 进行签名
	signature := o.sign(policyBase64)

	// 返回结果
	return &UploadPolicy{
		AccessKeyId: o.client.AccessKeyID,
		Policy:      policyBase64,
		Signature:   signature,
		Bucket:      o.BucketName,
		Endpoint:    o.Endpoint,
		Expire:      expireTime,
	}, nil
}

// 签名方法
func (o *OSSClient) sign(policyBase64 string) string {
	h := hmacSHA1([]byte(o.client.AccessKeySecret), []byte(policyBase64))
	return base64.StdEncoding.EncodeToString(h)
}

// HMAC-SHA1 签名工具
func hmacSHA1(key []byte, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	return h.Sum(nil)
}
