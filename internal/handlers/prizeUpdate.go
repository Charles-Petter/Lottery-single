package handlers

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/service"
	"time"

	"net/http"
)

type ViewPrize struct {
	Id           uint      `json:"id"`
	Title        string    `json:"title"`
	Img          string    `json:"img"`
	PrizeNum     int       `json:"prize_num"`
	PrizeCode    string    `json:"prize_code"`
	PrizeTime    uint      `json:"prize_time"`
	LeftNum      int       `json:"left_num"`
	PrizeType    uint      `json:"prize_type"`
	PrizePlan    string    `json:"prize_plan"`
	BeginTime    time.Time `json:"begin_time"`
	EndTime      time.Time `json:"end_time"`
	DisplayOrder uint      `json:"display_order"`
	SysStatus    uint      `json:"sys_status"`
}

func UpdatePrize(c *gin.Context) {
	var viewPrize ViewPrize
	if err := c.ShouldBindJSON(&viewPrize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	err := service.GetAdminService().UpdatePrize(c, (*service.ViewPrize)(&viewPrize))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prize updated successfully"})
}
