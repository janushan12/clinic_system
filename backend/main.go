package main

import (
	"log"
	"os"

	"clinic-notes/config"
	"clinic-notes/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()
	config.ConnectDB()
	defer config.DB.Close()

	r := gin.Default()

	// CORS — allow React dev server
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		// Core: AI parse + save
		api.POST("/parse", handlers.ParseAndSave)

		// Visits
		api.GET("/visits/:id", handlers.GetVisit)
		api.GET("/visits/:id/bill", handlers.GetBill)
		api.POST("/visits/:id/bill/pay", handlers.MarkBillPaid)
		api.GET("/visits/:id/pdf", handlers.DownloadPDF)

		// Patients
		api.GET("/patients", handlers.ListPatients)
		api.POST("/patients", handlers.CreatePatient)

		// Doctors
		api.GET("/doctors", handlers.ListDoctors)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}