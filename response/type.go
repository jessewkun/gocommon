package response

import "encoding/json"

// ApiResult 接口返回数据结构
type ApiResult struct {
	Code    int         `json:"code"`     // 接口错误码，0表示成功，非0表示失败
	Message string      `json:"message"`  // 提示信息
	Data    interface{} `json:"data"`     // 返回数据
	TraceId string      `json:"trace_id"` // 请求唯一标识
}

func (r *ApiResult) String() string {
	s, _ := json.Marshal(r)
	return string(s)
}
