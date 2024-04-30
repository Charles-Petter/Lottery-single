package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/service"
	"net/http"
)

type CouponImportHandler struct {
	req *service.ViewCouponInfo
	// resp     PrizeAddResponse
	resp HttpResponse

	// 需要什么Service，就在这里声明
	service service.AdminService
}

// CouponImport 获取奖品列表
func CouponImport(c *gin.Context) {
	// todo: 参数获取，校验
	h := CouponImportHandler{
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

func (h *CouponImportHandler) CheckInput(ctx context.Context) error {
	h.resp.Code = constant.ErrInputInvalid
	r := h.req
	fmt.Printf("Check input: %+v\n", r)
	if r.PrizeId <= 0 {
		log.ErrorContextf(ctx, "coupon import prize_id is invalid")
		return fmt.Errorf("coupon import prize_id is invalid")
	}
	if r.Code == "" {
		log.ErrorContextf(ctx, "coupon import code is invalid")
		return fmt.Errorf("coupon import code is invalid")
	}
	if r.SysStatus != 1 || r.SysStatus != 2 {
		log.ErrorContextf(ctx, "coupon import status is invalid")
		return fmt.Errorf("coupon import status is invalid")
	}
	h.resp.Code = constant.Success
	return nil
}

func (h *CouponImportHandler) Process(ctx context.Context) {
	log.Infof("CouponImportHandler req ==== %+v\n", h.req)
	successNum, failNum, err := h.service.ImportCoupon(ctx, h.req.PrizeId, h.req.Code)
	if err != nil {
		// TODO:
		h.resp.Code = constant.ErrInternalServer
		return
	}
	log.Infof("CouponImportHandler|successNum=%d|failNum=%d\n", successNum, failNum)
	// 继续处理
	return
}
