package oss

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
)

// STSConfig STS配置
type STSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	RoleArn         string
	SessionName     string
	Region          string
	Policy          string
}

// ResponseSTS STS参数
type ResponseSTS struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
	Expiration      string `json:"expiration"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	Callback        string `json:"callback"`
}

// GetSTS 获取STS参数
func GetSTS(s STSConfig) (*ResponseSTS, error) {
	client, err := sts.NewClientWithAccessKey(s.Region, s.AccessKeyID, s.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.RoleArn = s.RoleArn
	request.RoleSessionName = s.SessionName

	response, err := client.AssumeRole(request)
	if err != nil {
		return nil, err
	}
	return &ResponseSTS{
		AccessKeyID:     response.Credentials.AccessKeyId,
		AccessKeySecret: response.Credentials.AccessKeySecret,
		SecurityToken:   response.Credentials.SecurityToken,
		Expiration:      response.Credentials.Expiration,
		Bucket:          "",
		Endpoint:        "",
		Callback:        "",
	}, nil
}
