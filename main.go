package main

import (
	"luncher/handler/database"
	model "luncher/handler/models"
	"luncher/handler/utils"
	"luncher/service/telegramBot"

	"github.com/gin-gonic/gin"
)

func main() {

	utils.LoadENV()

	db := database.Connection()
	db.Conn.AutoMigrate(&model.Reserve{}, &model.User{}, &model.Meal{})

	app := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	telegramBot.LoadBot()

	telegramBot.StartBotServer()

	app.Run(":8085")
}
