package main

import (
	"log"
	"luncher/handler/database"
	model "luncher/handler/models"
	"luncher/handler/utils"
	"luncher/service/telegramBot"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	utils.LoadENV()

	db := database.Connection()
	db.Conn.AutoMigrate(&model.Reserve{}, &model.User{}, &model.Meal{})

	app := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	// Set timezone globally
	os.Setenv("TZ", "Asia/Tehran")                           // Change to your desired timezone
	time.Local = time.FixedZone("Tehran Time", 3*3600+30*60) // (UTC+3:30)

	log.Println(time.Now())

	telegramBot.LoadBot()

	telegramBot.StartBotServer()

	app.Run(":8085")
}
