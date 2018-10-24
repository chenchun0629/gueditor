package storage

import (
	"bytes"
	"io"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOssConfig struct {
	Appid     string
	AppSecret string
	Bucket    string
	EndPoint  string
	Domain    string // 外链域名
	IsPrivate bool   // 是否私有
	Expire    int64  // 有效时间
}

type AliyunOss struct {
	BaseInterface
	config *AliyunOssConfig // 配置信息
	client *oss.Client
	bucket *oss.Bucket
}

func NewAliyunOss(config *AliyunOssConfig) (*AliyunOss, error) {
	s := &AliyunOss{
		config: config,
	}

	var err error

	s.client, err = oss.New(s.config.EndPoint, s.config.Appid, s.config.AppSecret)

	if err != nil {
		return nil, err
	}

	s.bucket, err = s.client.Bucket(s.config.Bucket)

	if err != nil {
		return nil, err
	}

	return s, err
}

/**
从本地文件保存
*/
func (this *AliyunOss) SaveFileFromLocalPath(srcPath string, dstAbsPath, dstRelativePath string) (string, error) {
	fi, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	err = this.bucket.PutObject(dstRelativePath, fi)
	if err != nil {
		return "", err
	}

	return this.GetUrl(dstRelativePath)
}

func (this *AliyunOss) SaveFile(srcFile io.Reader, srcFileSize int64, dstAbsPath, dstRelativePath string) (string, error) {
	err := this.bucket.PutObject(dstRelativePath, srcFile)
	if err != nil {
		return "", err
	}

	return this.GetUrl(dstRelativePath)
}

func (this *AliyunOss) SaveData(data []byte, dstAbsPath, dstRelativePath string) (string, error) {
	err := this.bucket.PutObject(dstRelativePath, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return this.GetUrl(dstRelativePath)
}

func (this *AliyunOss) GetUrl(objectKey string) (string, error) {
	if objectKey == "" {
		return "", nil
	}

	if this.config.IsPrivate {
		if this.config.Expire > 0 {
			return this.bucket.SignURL(objectKey, oss.HTTPGet, this.config.Expire)
		} else {
			return this.bucket.SignURL(objectKey, oss.HTTPGet, 604800) // 7天
		}
	} else {
		return this.config.Domain + objectKey, nil
	}
}
