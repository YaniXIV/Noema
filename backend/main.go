package main

import (
	"fmt"
	"github.com/joho/godotenv"
	//s "noema/server"
	"context"
	"google.golang.org/genai"
	"log"
)

func initVariables() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func makeQuery(s string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text(s),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Text())

}

func main() {
	initVariables()
	var userin string
	for {
		fmt.Scanln(&userin)
		makeQuery(userin)
	}
}
