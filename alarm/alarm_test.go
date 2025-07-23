package alarm

import (
	"context"
	"testing"
)

func TestSendAlarm_UnifiedInterface(t *testing.T) {
	Cfg = &Config{
		Bark: &Bark{
			BarkIds: []string{"jT64URJj8b6Fp9Y3nVKJiP"},
		},
		Feishu: &Feishu{
			WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/",
			Secret:     "",
		},
		Timeout: 10,
	}

	// 初始化
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 测试统一接口发送报警
	title := "统一报警测试"
	content := []string{
		"- 第一行内容",
		"- 第二行内容",
		"- 第三行内容",
	}

	err := SendAlarm(context.Background(), title, content)
	if err != nil {
		t.Fatalf("SendAlarm failed: %v", err)
	}
}
