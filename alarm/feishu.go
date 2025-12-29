package alarm

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/http"
)

type Feishu struct {
	WebhookURL string `mapstructure:"webhook_url" json:"webhook_url"` // Feishu 机器人 Webhook URL
	Secret     string `mapstructure:"secret" json:"secret"`           // Feishu 机器人 Secret
}

// FeishuMessage 飞书消息结构
type FeishuMessage struct {
	Timestamp int64       `json:"timestamp"`
	Sign      string      `json:"sign"`
	MsgType   string      `json:"msg_type"`
	Content   interface{} `json:"content"`
}

// FeishuPostContent 飞书富文本消息内容
type FeishuPostContent struct {
	Post *FeishuPost `json:"post"`
}

// FeishuPost 飞书富文本消息
type FeishuPost struct {
	ZhCn *FeishuPostLang `json:"zh_cn"`
}

// FeishuPostLang 飞书富文本消息语言版本
type FeishuPostLang struct {
	Title   string            `json:"title"`
	Content [][]FeishuElement `json:"content"`
}

// FeishuElement 飞书富文本消息元素
type FeishuElement struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
	Href string `json:"href,omitempty"`
}

// Send 发送飞书消息
func (f *Feishu) Send(ctx context.Context, title string, content []string) error {
	if f.WebhookURL == "" {
		return fmt.Errorf("feishu webhook URL is not configured")
	}

	// 构建富文本内容
	var contentRows [][]FeishuElement
	for _, line := range content {
		contentRows = append(contentRows, []FeishuElement{{
			Tag:  "text",
			Text: line,
		}})
	}

	// 生成签名
	timestamp := time.Now().Unix()
	signature, err := genSign(f.Secret, timestamp)
	if err != nil {
		return fmt.Errorf("failed to generate signature: %v", err)
	}

	// 构建消息结构
	message := FeishuMessage{
		Timestamp: timestamp,
		Sign:      signature,
		MsgType:   "post",
		Content: FeishuPostContent{
			Post: &FeishuPost{
				ZhCn: &FeishuPostLang{
					Title:   title,
					Content: contentRows,
				},
			},
		},
	}

	// 序列化消息
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal feishu message: %w", err)
	}

	// 使用 @http 模块发送请求
	req := http.RequestPost{
		URL:     f.WebhookURL,
		Payload: payload, // resty 会自动处理序列化，但我们已手动序列化以构建完整FeishuMessage，这里传递字节
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	resp, err := alarmHTTPClient.Post(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send feishu request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code from feishu: %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}

// genSign 生成 Feishu 签名
func genSign(secret string, timestamp int64) (string, error) {
	if len(secret) == 0 {
		return "", nil
	}
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
