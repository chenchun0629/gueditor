package storage

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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

	objectKey := this.GetName(dstRelativePath)
	err = this.bucket.PutObject(objectKey, fi)
	if err != nil {
		return "", err
	}

	return this.GetUrl(objectKey)
}

func (this *AliyunOss) SaveFile(srcFile io.Reader, srcFileSize int64, dstAbsPath, dstRelativePath string) (string, error) {
	objectKey := this.GetName(dstRelativePath)
	err := this.bucket.PutObject(objectKey, srcFile)
	if err != nil {
		return "", err
	}

	return this.GetUrl(objectKey)
}

func (this *AliyunOss) SaveData(data []byte, dstAbsPath, dstRelativePath string) (string, error) {
	var ext string
	ct := http.DetectContentType(data)
	if ct == "application/octet-stream" {
		ext = filepath.Ext(dstRelativePath)
	} else {
		ext = "." + strings.Split(ct, "/")[1]
	}

	objectKey := this.GetNameWithExt(dstRelativePath, ext)
	err := this.bucket.PutObject(objectKey, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return this.GetUrl(objectKey)
}

func (this *AliyunOss) GetUrl(objectKey string) (string, error) {
	if objectKey == "" {
		return "", nil
	}

	ext := filepath.Ext(objectKey)

	if this.config.IsPrivate {
		if this.config.Expire > 0 {
			return this.bucket.SignURL(objectKey, oss.HTTPGet, this.config.Expire)
		} else {
			return this.bucket.SignURL(objectKey, oss.HTTPGet, 604800) // 7天
		}
	} else {
		if strings.ToLower(ext) == ".webp" {
			return this.config.Domain + objectKey + "?x-oss-process=image/format,png", nil
		} else {
			return this.config.Domain + objectKey, nil
		}
	}
}

func (this *AliyunOss) GetNameWithExt(dstRelativePath, ext string) string {
	objectKey := this.Uuid()
	if ext != "" {
		objectKey = objectKey + ext
	}
	return objectKey
}

func (this *AliyunOss) GetName(dstRelativePath string) string {
	ext := path.Ext(dstRelativePath)
	return this.GetNameWithExt(dstRelativePath, ext)
}

func (this *AliyunOss) Uuid() string {
	unix32bits := uint32(time.Now().UTC().Unix())

	buff := make([]byte, 12)

	rand.Read(buff)

	return fmt.Sprintf("%x-%x-%x-%x-%x-%x", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:])
}
