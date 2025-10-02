package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"spendingmanagement/server/auth"
	"spendingmanagement/server/db"
	"spendingmanagement/server/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// --- Load configuration ---
	port := getenvDefault("PORT", "8080")
	projectID := getenvDefault("FIREBASE_PROJECT_ID", "")
	credFile := getenvDefault("GOOGLE_APPLICATION_CREDENTIALS", "") // path to serviceAccountKey.json

	if projectID == "" || credFile == "" {
		log.Println("[WARN] FIREBASE_PROJECT_ID or GOOGLE_APPLICATION_CREDENTIALS not set.")
		log.Println("       Token verification will likely fail. Set env vars before running in production.")
	}

	// --- Init Firebase Auth client ---
	authClient, err := auth.NewFirebaseAuthClient(context.Background(), projectID, credFile)
	if err != nil {
		log.Fatalf("failed to init firebase auth: %v", err)
	}

	// --- Init DB ---
	dbConn, err := db.Open()
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	if err := dbConn.AutoMigrate(&models.Spending{}); err != nil {
		log.Fatalf("failed to migrate DB: %v", err)
	}

	// --- Router ---
	r := gin.Default()

	// Health
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "time": time.Now()})
	})

	// Public ping
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "SpendingManagement API (Go + Firebase)"})
	})

	// Protected group
	authMw := auth.AuthMiddleware(authClient)
	api := r.Group("/api", authMw)

	api.GET("/me", func(c *gin.Context) {
		claims := auth.GetClaims(c) // contains UID, Email, etc.
		c.JSON(http.StatusOK, claims)
	})

	// Spendings CRUD (scoped to authenticated user)
	api.GET("/spendings", func(c *gin.Context) {
		uid := auth.MustUID(c)
		var items []models.Spending
		if err := dbConn.Where("user_id = ?", uid).Order("date desc, id desc").Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	})

	type createReq struct {
		Amount    float64   `json:"amount" binding:"required"`
		Category  string    `json:"category" binding:"required"`
		Note      string    `json:"note"`
		Date      time.Time `json:"date"`
		ImageURL  string    `json:"image_url"`
		Currency  string    `json:"currency"`
	}

	api.POST("/spendings", func(c *gin.Context) {
		uid := auth.MustUID(c)
		var req createReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item := models.Spending{
			UserID:   uid,
			Amount:   req.Amount,
			Category: req.Category,
			Note:     req.Note,
			Date:     defaultTime(req.Date),
			ImageURL: req.ImageURL,
			Currency: firstNonEmpty(req.Currency, "VND"),
		}
		if err := dbConn.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	})

	api.PUT("/spendings/:id", func(c *gin.Context) {
		uid := auth.MustUID(c)
		id := c.Param("id")
		var existing models.Spending
		if err := dbConn.Where("id = ? AND user_id = ?", id, uid).First(&existing).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		var req createReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		existing.Amount = req.Amount
		existing.Category = req.Category
		existing.Note = req.Note
		if !req.Date.IsZero() {
			existing.Date = req.Date
		}
		existing.ImageURL = req.ImageURL
		if req.Currency != "" {
			existing.Currency = req.Currency
		}
		if err := dbConn.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
	})

	api.DELETE("/spendings/:id", func(c *gin.Context) {
		uid := auth.MustUID(c)
		id := c.Param("id")
		if err := dbConn.Where("id = ? AND user_id = ?", id, uid).Delete(&models.Spending{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": id})
	})

	log.Printf("listening on :%s ...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func firstNonEmpty(a, b string) string {
	if a != "" { return a }
	return b
}

func defaultTime(t time.Time) time.Time {
	if t.IsZero() { return time.Now() }
	return t
}
