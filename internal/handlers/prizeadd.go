package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/service"
	"net/http"
	"time"
)

type PrizeAddHandler struct {
	req *service.ViewPrize
	// resp     PrizeAddResponse
	resp *HttpResponse

	// 需要什么Service，就在这里声明
	service service.AdminService
}

// PrizeAdd 获取奖品列表
func PrizeAdd(c *gin.Context) {
	// todo: 参数获取，校验
	h := PrizeAddHandler{
		req:     &service.ViewPrize{},
		resp:    &HttpResponse{},
		service: service.GetAdminService(),
	}
	// HTTP响应
	defer func() {
		// 通过对应的Code，获取Msg
		h.resp.Msg = constant.GetErrMsg(h.resp.Code)
		c.JSON(http.StatusOK, h.resp)
	}()
	// 获取请求数据
	if err := c.ShouldBind(&h.req); err != nil {
		fmt.Printf("Error binding:%v", err)
		h.resp.Code = constant.ErrShouldBind
		return
	}
	Run(&h)
}

func (h *PrizeAddHandler) CheckInput(ctx context.Context) error {
	h.resp.Code = constant.ErrInputInvalid
	r := h.req
	fmt.Printf("Check input: %+v\n", r)
	if r.Title == "" {
		log.ErrorContextf(ctx, "prize title is invalid")
		return fmt.Errorf("prize title is invalid")
	}
	if r.Img == "" {
		log.ErrorContextf(ctx, "prize img is invalid")
		return fmt.Errorf("prize img is invalid")
	}
	if r.PrizeNum <= 0 {
		log.ErrorContextf(ctx, "prize prize_num is invalid")
		return fmt.Errorf("prize prize_num is invalid")
	}
	if r.PrizeCode == "" {
		log.ErrorContextf(ctx, "prize prize_code is invalid")
		return fmt.Errorf("prize prize_code is invalid")
	}
	if r.EndTime.Before(time.Now()) {
		log.ErrorContextf(ctx, "prize end_time is invalid")
		return fmt.Errorf("prize end_time is invalid")
	}
	if r.PrizeType > constant.PrizeTypeEntityLarge {
		log.ErrorContextf(ctx, "prize prize_type is invalid")
		return fmt.Errorf("prize prize_type is invalid")
	}
	h.resp.Code = constant.Success
	return nil
}

func (h *PrizeAddHandler) Process(ctx context.Context) {
	log.Infof("PrizeAddHandler req ==== %+v\n", h.req)
	err := h.service.AddPrize(ctx, h.req)
	if err != nil {
		// TODO:
		h.resp.Code = constant.ErrInternalServer
		// log.Errorf()
		return
	}

	// 继续处理
	return
}
