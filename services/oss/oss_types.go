package oss

import (
	"encoding/xml"
	"time"
)

// Object 定义 OSS 对象
type Object struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         int64     `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
}

// ListBucketResult 列出 Bucket 中的对象返回值
type ListBucketResult struct {
	XMLName  xml.Name `xml:"ListBucketResult"`
	Name     string   `xml:"Name"`
	Prefix   string   `xml:"Prefix"`
	Marker   string   `xml:"Marker"`
	MaxKeys  int      `xml:"MaxKeys"`
	Contents []Object `xml:"Contents"`
}
