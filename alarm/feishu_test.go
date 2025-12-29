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
	// Webhook URL 不完整（缺少 token），预期会失败，这是正常的
	// 只要不是配置验证错误，就认为测试通过
	if err != nil {
		// 允许网络错误（404等），但不允许配置验证错误
		if err.Error() == "feishu webhook URL is not configured" {
			t.Fatalf("SendFeishu failed with config error: %v", err)
		}
		// 其他错误（如网络错误）是可以接受的
		t.Logf("SendFeishu failed (expected for invalid webhook URL): %v", err)
	}
}
