package handlers

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/service"
	"net/http"
	"strconv"
)

func DeletePrize(c *gin.Context) {
	prizeIDStr := c.Param("id")
	prizeID, err := strconv.ParseUint(prizeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prize ID"})
		return
	}

	err = service.GetAdminService().DeletePrize(c, uint(prizeID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prize deleted successfully"})
}
