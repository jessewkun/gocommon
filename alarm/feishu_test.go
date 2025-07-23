package alarm

import (
	"context"
	"testing"
)

func TestFeishu_Send(t *testing.T) {
	f := &Feishu{
		WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/", // 带参数测试签名拼接
		Secret:     "",
	}

	// 设置全局Cfg，避免空指针
	Cfg.Feishu = f
	Cfg.Timeout = 10

	err := f.Send(context.Background(), "单元测试标题", []string{"- 第一行内容", "- 第二行内容"})
	if err != nil {
		t.Fatalf("SendFeishu failed: %v", err)
	}
}
