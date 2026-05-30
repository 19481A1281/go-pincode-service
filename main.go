package main

import (
	"log"
	"os"

	"github.com/19481A1281/go-pincode-service/connection"
	"github.com/19481A1281/go-pincode-service/controllers"
	"github.com/19481A1281/go-pincode-service/excel"
	"github.com/19481A1281/go-pincode-service/models"
	"github.com/19481A1281/go-pincode-service/repositories"
	"github.com/19481A1281/go-pincode-service/routes"
	"github.com/19481A1281/go-pincode-service/services"
	"github.com/gin-gonic/gin"
)

func main() {
	connection, err := connection.NewMysqlConnection()
	if err!=nil{
		log.Fatal("Unable to establish db connection",err)
	}

	// defer connection.Close()
	connection.DB.AutoMigrate(&models.Pincode{})

	// Initialize background CSV loading
	excel.InitCSVData()

	repo := repositories.NewPincodeRepository(connection.DB)
	service := services.NewPincodeService(repo)
	contoller := controllers.NewPincodeController(service)

	router := gin.Default()
	routes.SetupRoutes(router, contoller)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
