package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"lottery_single/internal/handlers/params"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/service"
	"net/http"
)

type LoginHandler struct {
	req *params.LoginReq
	// resp     PrizeListResponse
	resp *HttpResponse

	// 需要什么Service，就在这里声明
	service service.UserService
}

// Login 用户登录
func Login(c *gin.Context) {
	// todo: 参数获取，校验
	h := LoginHandler{
		req:     &params.LoginReq{},
		resp:    &HttpResponse{},
		service: service.GetUserService(),
	}
	// HTTP响应
	defer func() {
		// 通过对应的Code，获取Msg
		h.resp.Msg = constant.GetErrMsg(h.resp.Code)
		c.JSON(http.StatusOK, h.resp)
	}()
	// 获取请求数据
	if err := c.ShouldBind(h.req); err != nil {
		log.Errorf("ShouldBind login req:err+%v\n", err)
		return
	}
	log.Infof("login req:%v\n", h.req)
	Run(&h)
}

func (l *LoginHandler) CheckInput(ctx context.Context) error {
	if err := recover(); err != nil {
		fmt.Printf("CheckInput|login panic:%+v\n", err)
	}
	if l.req == nil {
		l.resp.Code = constant.ErrInputInvalid
		log.Errorf("login params is nil")
		return fmt.Errorf(constant.GetErrMsg(constant.ErrInputInvalid))
	}
	if l.req.UserName == "" || l.req.PassWord == "" {
		l.resp.Code = constant.ErrInputInvalid
		log.Errorf("login params invalid, user_name=%s,pass_word=%s\n", l.req.UserName, l.req.PassWord)
		return fmt.Errorf(constant.GetErrMsg(constant.ErrInputInvalid))
	}
	return nil
}

func (l *LoginHandler) Process(ctx context.Context) {
	v, err := l.service.Login(ctx, l.req.UserName, l.req.PassWord)
	if err != nil {
		log.ErrorContextf(ctx, "LoginHandler|process login err:%v", err)
		l.resp.Code = constant.ErrLogin
		return
	}
	// 继续处理
	l.resp.Data = v
	return
}
