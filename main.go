package main

import (
	"fmt"
	"github.com/godrealms/ali-fusion-go-sdk/config"
	"github.com/godrealms/ali-fusion-go-sdk/core"
	"github.com/godrealms/ali-fusion-go-sdk/services/oss"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化核心客户端
	client := core.NewClient(cfg.AccessKeyID, cfg.AccessKeySecret, cfg.Region)

	// 初始化 OSS 客户端
	ossClient := oss.NewOSSClient(client, "my-bucket", "oss-cn-hangzhou.aliyuncs.com")

	// 上传文件
	path, err := ossClient.UploadFile("example.txt", "./example.txt")
	if err != nil {
		fmt.Println("Upload failed:", err)
		return
	}
	fmt.Println("File uploaded successfully: ", path)

	// 列出文件
	objects, err := ossClient.ListObjects("")
	if err != nil {
		fmt.Println("List objects failed:", err)
		return
	}
	fmt.Println("Objects in bucket:")
	for _, obj := range objects {
		fmt.Printf("- %s (Size: %d)\n", obj.Key, obj.Size)
	}

	// 下载文件
	err = ossClient.DownloadFile("example.txt", "./downloaded_example.txt")
	if err != nil {
		fmt.Println("Download failed:", err)
		return
	}
	fmt.Println("File downloaded successfully!")

	// 删除文件
	err = ossClient.DeleteObject("example.txt")
	if err != nil {
		fmt.Println("Delete failed:", err)
		return
	}
	fmt.Println("File deleted successfully!")
}
