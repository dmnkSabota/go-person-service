package main

import (
	"log"
	"os"
	"person-service/database"
	"person-service/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting Person Service...")

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")

	personHandler := handlers.NewPersonHandler(db)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/save", personHandler.SavePerson)
	router.GET("/:id", personHandler.GetPerson)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
