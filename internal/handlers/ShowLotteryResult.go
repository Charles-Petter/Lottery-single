package handlers

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/repo"
	"net/http"
)

func ShowLotteryResult(c *gin.Context) {
	resultRepo := repo.NewResultRepo()
	results, err := resultRepo.GetAll(gormcli.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve lottery results"})
		return
	}
	c.JSON(http.StatusOK, results)
}
