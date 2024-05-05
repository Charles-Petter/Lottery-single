package handlers

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/model"
	"lottery_single/internal/service"
	"net/http"
	"strconv"
)

// 添加IP到黑名单
func AddBlackIP(c *gin.Context) {
	var blackIP model.BlackIp
	if err := c.ShouldBindJSON(&blackIP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := service.GetAdminService().AddBlackIP(c, &blackIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP added to blacklist successfully"})
}

// 删除黑名单中的IP
func DeleteBlackIP(c *gin.Context) {
	id := c.Param("id")
	blackIpID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = service.GetAdminService().DeleteBlackIP(c, uint(blackIpID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP removed from blacklist successfully"})
}

// 查看所有黑名单IP
func ListBlackIP(c *gin.Context) {
	blackIPs, err := service.GetAdminService().GetAllBlackIPs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, blackIPs)
}
