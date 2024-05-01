package handlers

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/service"
	"net/http"
	"strconv"
)

func AddUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Errorf("AddUser: error binding user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := service.GetAdminService().AddUser(c.Request.Context(), &user); err != nil {
		log.Errorf("AddUser: error adding user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user added successfully"})
}

func UpdateUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("userID"), 10, 64)
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Errorf("UpdateUser: error binding user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := service.GetAdminService().UpdateUser(c.Request.Context(), uint(userID), &user); err != nil {
		log.Errorf("UpdateUser: error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func DeleteUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("userID"), 10, 64)

	if err := service.GetAdminService().DeleteUser(c.Request.Context(), uint(userID)); err != nil {
		log.Errorf("DeleteUser: error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func GetAllUsers(c *gin.Context) {
	users, err := service.GetAdminService().GetAllUsers(c.Request.Context())
	if err != nil {
		log.Errorf("GetAllUsers: error getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
