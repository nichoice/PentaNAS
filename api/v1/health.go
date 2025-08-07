package v1

// PingResponse 定义 ping 接口的响应结构
type PingResponse struct {
	Message string `json:"message" example:"pong"`
}