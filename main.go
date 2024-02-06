package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"telegram-ki-maya/api"
	"telegram-ki-maya/pkg"
	"telegram-ki-maya/subscription"
)

func main() {
	app := fiber.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := pkg.Connect(os.Getenv("TOKEN"), true)
	server := subscription.NewServer(client)
	sub := subscription.NewService()
	sub.SetServer(server)
	sub.Run()

	app.Get("/connect", api.Connect(sub))
	log.Fatalln(app.Listen(":8060"))
}