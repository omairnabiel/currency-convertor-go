package main

import (
	"github.com/gin-gonic/gin"
	process "zlack-home/process"
)

func main() {
	// Init router
	router := gin.New()

	router.GET("/process-transactions", process.Process)

	router.Run(":8080")
}