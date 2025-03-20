package main

import (
	"fmt"
	"log"
	"os"

	ps "screenGPT/proxy-service-microservice"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	defaultModel     = "gpt-4-vision-preview"
	defaultMaxTokens = 300
)

func main() {
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Fatalf("OPENAI_API_KEY environment variable not set")
		return
	}

	controller := ps.NewController(
		openAIAPIKey,
		"https://api.proxyapi.ru",
		defaultModel,
		defaultMaxTokens,
	)

	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/upload", controller.UploadHandler)

	port := 8080
	log.Printf("Server listening on port %d\n", port)
	log.Fatal(router.Run(fmt.Sprintf(":%d", port)))
}
