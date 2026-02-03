package main

import (
	"fmt"
	"github.com/joho/godotenv"
	//s "noema/server"
	"bufio"
	"context"
	"google.golang.org/genai"
	"log"
	"os"
)

func initVariables() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func makeQuery(s string, model string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Models.GenerateContent(
		ctx,
		model,
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
	modelName := "gemini-3-flash-preview" // Use a Gemini 3 model name
	//for {
	fmt.Println("Enter Query")
	reader := bufio.NewReader(os.Stdin)
	userin, _ := reader.ReadString('\n')

	makeQuery(userin, modelName)
	//}
}
