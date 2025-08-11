package tests

import (
	"bytes"
	"encoding/json"
	"entain-task/internal/db"
	"entain-task/internal/handlers"
	"entain-task/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	var err error
	db.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory: %v", err)
	}
	if err := models.AutoMigrate(db.DB); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	db.DB.Create(&models.User{ID: 1, Balance: 0.00})
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/user/:userId/transaction", handlers.PostTransaction)
	r.GET("/user/:userId/balance", handlers.GetBalance)
	return r
}

func TestTransactionFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)
	router := setupRouter()

	// Check initial balance
	req := httptest.NewRequest("GET", "/user/1/balance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var balanceResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &balanceResp)
	if balanceResp["balance"] != "0.00" {
		t.Fatalf("expected balance 0.00, got %v", balanceResp["balance"])
	}

	// Win transaction
	body := `{"state":"win","amount":"10.15","transactionId":"tx-001"}`
	req = httptest.NewRequest("POST", "/user/1/transaction", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Check balance again
	req = httptest.NewRequest("GET", "/user/1/balance", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &balanceResp)
	if balanceResp["balance"] != "10.15" {
		t.Fatalf("expected balance 10.15, got %v", balanceResp["balance"])
	}

	// Lose transaction
	body = `{"state":"lose","amount":"1.15","transactionId":"tx-002"}`
	req = httptest.NewRequest("POST", "/user/1/transaction", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Balance after lose
	req = httptest.NewRequest("GET", "/user/1/balance", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &balanceResp)
	if balanceResp["balance"] != "9.00" {
		t.Fatalf("expected balance 9.00, got %v", balanceResp["balance"])
	}

	// Duplicate transaction
	body = `{"state":"win","amount":"5.00","transactionId":"tx-002"}`
	req = httptest.NewRequest("POST", "/user/1/transaction", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate transaction, got %d", w.Code)
	}

	// Negative balance attempt
	body = `{"state":"lose","amount":"9999.99","transactionId":"tx-003"}`
	req = httptest.NewRequest("POST", "/user/1/transaction", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 negative balance, got %d", w.Code)
	}
}
