package response

import "encoding/json"

// APIResult 接口返回数据结构
type APIResult struct {
	Code    int         `json:"code"`     // 接口错误码，0表示成功，非0表示失败
	Message string      `json:"message"`  // 提示信息
	Data    interface{} `json:"data"`     // 返回数据
	TraceID string      `json:"trace_id"` // 请求唯一标识
}

func (r *APIResult) String() string {
	s, _ := json.Marshal(r)
	return string(s)
}
