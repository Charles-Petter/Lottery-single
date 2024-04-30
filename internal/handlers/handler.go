package handlers

import (
	"context"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/utils"
)

// HttpResponse http独立请求返回结构体,这个通用的，不需要修改
type HttpResponse struct {
	Code constant.ErrCode `json:"code"`
	Msg  string           `json:"msg"`
	Data interface{}      `json:"data"`
}

type Handler interface {
	CheckInput(ctx context.Context) error
	Process(ctx context.Context)
}

// Run 执行函数
func Run(handler Handler) {
	ctx := context.WithValue(context.Background(), constant.ReqID, utils.NewUuid())
	// 1. 参数校验
	err := handler.CheckInput(ctx)
	// 校验失败，
	if err != nil {
		return
	}
	// 2. 逻辑处理
	handler.Process(ctx)
}
