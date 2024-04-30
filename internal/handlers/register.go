package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"lottery_single/internal/handlers/params"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/service"
	"net/http"
)

type RegisterHandler struct {
	req     *params.RegisterReq
	resp    *HttpResponse
	service service.UserService
}

func Register(c *gin.Context) {
	h := RegisterHandler{
		req:     &params.RegisterReq{},
		resp:    &HttpResponse{},
		service: service.GetUserService(),
	}

	defer func() {
		h.resp.Msg = constant.GetErrMsg(h.resp.Code)
		c.JSON(http.StatusOK, h.resp)
	}()

	if err := c.ShouldBind(h.req); err != nil {
		log.Errorf("ShouldBind register req:err+%v\n", err)
		return
	}
	log.Infof("register req:%v\n", h.req)
	Run(&h)
}
func (r *RegisterHandler) CheckInput(ctx context.Context) error {
	if r.req == nil {
		r.resp.Code = constant.ErrInputInvalid
		log.Errorf("register params are nil")
		return fmt.Errorf(constant.GetErrMsg(constant.ErrInputInvalid))
	}

	// 检查用户名、密码、电子邮件、手机号、真实姓名、年龄和性别
	if r.req.UserName == "" || r.req.PassWord == "" || r.req.Email == "" || r.req.Mobile == "" ||
		r.req.RealName == "" || r.req.Age < 0 || (r.req.Gender != "male" && r.req.Gender != "female") {
		r.resp.Code = constant.ErrInputInvalid
		log.Errorf("register params are invalid: %+v", r.req)
		return fmt.Errorf(constant.GetErrMsg(constant.ErrInputInvalid))
	}

	return nil
}

func (r *RegisterHandler) Process(ctx context.Context) {
	// 将请求参数封装成 *model.User 类型的对象
	newUser := &model.User{
		UserName: r.req.UserName,
		Password: r.req.PassWord,
		Email:    r.req.Email,
		Mobile:   r.req.Mobile,
		RealName: r.req.RealName,
		Age:      r.req.Age,
		Gender:   r.req.Gender,
	}

	// 调用 service.Register 方法进行注册
	err := r.service.Register(ctx, newUser)
	if err != nil {
		log.ErrorContextf(ctx, "RegisterHandler|process register err: %v", err)
		r.resp.Code = constant.ErrRegister
		return
	}

	// 处理注册成功后的结果（根据具体需求进行调整）
	r.resp.Data = newUser
}
