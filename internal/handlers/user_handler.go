package handlers

import (
	"entain-task/internal/db"
	"entain-task/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type transactionRequest struct {
	State         string `json:"state" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	TransactionID string `json:"transactionId" binding:"required"`
}

type balanceResponse struct {
	UserID  uint64 `json:"userId"`
	Balance string `json:"balance"`
}

func PostTransaction(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil || userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}

	sourceType := c.GetHeader("Source-Type")
	if sourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing Source-Type header"})
		return
	}

	var req transactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = models.UpdateUserBalance(userId, req.TransactionID, req.State, req.Amount)
	if err != nil {
		switch err {
		case models.ErrDuplicateTransaction:
			c.JSON(http.StatusBadRequest, gin.H{"error": "transaction already processed"})
		case models.ErrNegativeBalance:
			c.JSON(http.StatusBadRequest, gin.H{"error": "balance cannot be negative"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func GetBalance(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil || userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, balanceResponse{
		UserID:  user.ID,
		Balance: strconv.FormatFloat(user.Balance, 'f', 2, 64),
	})
}
