package telegramBot

import (
	"encoding/json"
	"fmt"
	"log"
	"luncher/handler/database"
	model "luncher/handler/models"
	"luncher/handler/utils"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/gorm"
)

var telegramBot *tgbotapi.BotAPI

func LoadBot() {

	// Replace with your Bot's token
	botToken := utils.Getenv("TELEGRAM_BOT_TOKEN", "")

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	telegramBot = bot
}

func StartBotServer() {

	// Handle updates (messages, button presses, etc.)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := telegramBot.GetUpdatesChan(u)

	db := database.Connection().Conn

	// Loop to listen for incoming messages or button presses
	for update := range updates {
		if update.Message != nil {

			// Start command to show the meal selection form
			if update.Message.Text == "/start" {

				user := model.User{
					TelegramID: update.Message.Chat.ID,
					Username:   update.Message.Chat.UserName,
					Name:       update.Message.Chat.FirstName,
				}

				var userDB model.User
				db.Where("telegram_id = ?", user.TelegramID).First(&userDB)
				if userDB.ID == 0 {
					db.Create(&user)
				}

				// todo: select always lunch/dinner

				// welcome message
				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "سلام"))
			}

			user := findUser(db, update.Message.Chat.ID)

			if user.ID == 0 {
				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در ارتباط با دیتابیس"))
				continue
			}

			if update.Message.Text == "/select" {

				showMealSelectionForm(user, update.Message.Chat.ID)
			}

		}

		// Handle button presses (callback queries)
		if update.CallbackQuery != nil {

			//find user id

			user := findUser(db, int64(update.CallbackQuery.From.ID))

			if user.ID == 0 {
				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در ارتباط با دیتابیس"))
				continue
			}

			handleButtonPress(user, update.CallbackQuery)
		}
	}
}

func findUser(db *gorm.DB, id int64) model.User {
	user := model.User{}

	err := db.Model(&model.User{}).Where("telegram_id = ?", id).First(&user)
	if err.Error != nil {
		log.Println(err.Error)
	}

	return user
}

// Show the meal selection form with inline buttons
func showMealSelectionForm(user model.User, chatID int64) {

	buttons := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("همه شام ها", "all_dinner"),
			tgbotapi.NewInlineKeyboardButtonData("همه نهار ها", "all_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("*", "all"),
		),
	}

	var next14DaysReserves []model.Reserve
	db := database.Connection().Conn
	db.Where("user_id = ? AND date >= ?", user.ID, time.Now().Truncate(24*time.Hour)).Find(&next14DaysReserves)

	meals := generateMeals()

	for i := 0; i < 14; i++ {

		date := time.Now().AddDate(0, 0, i)

		// find week number of year
		weekNumber := utils.GetWeekNumber(date) % 2
		weekDay := date.Weekday()
		faDayName := utils.GetFaDayName(weekDay)

		var lunchMeals []string
		var dinnerMeals []string

		dayIndex := utils.GetJalaliWeekDayNumber(weekDay) - 1
		if weekNumber == 0 {
			lunchMeals = meals["firstLunchMeals"]
			dinnerMeals = meals["firstDinnerMeals"]
		} else {
			lunchMeals = meals["secondLunchMeals"]
			dinnerMeals = meals["secondDinnerMeals"]
		}

		// Check if the user has already selected a meal for this day
		var selectedMeal model.Reserve
		for _, reserve := range next14DaysReserves {

			if reserve.Date.Format("2006-01-02") == date.Format("2006-01-02") {
				selectedMeal = reserve
				break
			}
		}

		rowButton := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(lunchMeals[dayIndex], selectedMeal.HasDinner), fmt.Sprintf("%s_dinner", date.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(dinnerMeals[dayIndex], selectedMeal.HasLunch), fmt.Sprintf("%s_lunch", date.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(faDayName, date.Format("2006-01-02")),
		)

		buttons = append(buttons, rowButton)
	}

	// Create inline keyboard buttons for each day and meal (lunch and dinner)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "Please select your meal preferences for each day.")
	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = true
	telegramBot.Send(msg)
}

func generateMeals() map[string][]string {

	firstWeekLunch := utils.Getenv("FIRST_WEEK_LUNCH", "")
	secondWeekLunch := utils.Getenv("SECOND_WEEK_LUNCH", "")
	firstWeekDinner := utils.Getenv("FIRST_WEEK_DINNER", "")
	secondWeekDinner := utils.Getenv("SECOND_WEEK_DINNER", "")

	var firstLunchMeals []string
	var secondLunchMeals []string
	var firstDinnerMeals []string
	var secondDinnerMeals []string

	err := json.Unmarshal([]byte(firstWeekLunch), &firstLunchMeals)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal([]byte(secondWeekLunch), &secondLunchMeals)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal([]byte(firstWeekDinner), &firstDinnerMeals)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal([]byte(secondWeekDinner), &secondDinnerMeals)
	if err != nil {
		log.Println(err)
	}

	return map[string][]string{
		"firstLunchMeals":   firstLunchMeals,
		"secondLunchMeals":  secondLunchMeals,
		"firstDinnerMeals":  firstDinnerMeals,
		"secondDinnerMeals": secondDinnerMeals,
	}
}

// Handle button press events
func handleButtonPress(user model.User, callback *tgbotapi.CallbackQuery) {

	// Get the user ID and the selected meal option
	selectedOption := callback.Data

	var isLunch bool
	var isDinner bool

	if selectedOption == "all" || selectedOption == "all_lunch" || selectedOption == "all_dinner" {
		for i := 0; i < 14; i++ {
			date := time.Now().AddDate(0, 0, i).Truncate(24 * time.Hour)
			db := database.Connection().Conn

			var reserve model.Reserve

			query := db.Model(&model.Reserve{})
			query.Where("date = ? ", date)
			query.Where("user_id = ?", user.ID)
			err := query.First(&reserve).Error

			if err != nil && err != gorm.ErrRecordNotFound {
				log.Println(err)
				return
			}

			if err == gorm.ErrRecordNotFound {

				reserve = model.Reserve{
					Date:   date,
					UserID: user.ID,
				}
			}

			if selectedOption == "all" {
				reserve.HasLunch = true
				reserve.HasDinner = true
			} else if selectedOption == "all_lunch" {
				reserve.HasLunch = true
			} else if selectedOption == "all_dinner" {
				reserve.HasDinner = true
			}

			reserve.UpdateAt = time.Now()

			db.Save(&reserve)
		}

		// Send the updated message to the user
		telegramBot.AnswerCallbackQuery(tgbotapi.NewCallback(callback.ID, "همه انتخاب شدند"))

	} else {

		// if end with _lunch or _dinner
		// split and get date
		date, err := time.Parse("2006-01-02", selectedOption[:10])
		if err != nil {
			log.Println(err)
			return
		}

		// if date is today and hour pass from 20 in tehran
		// show error message
		location, _ := time.LoadLocation("Asia/Tehran")
		if date.Before(time.Now().AddDate(0, 0, 1).Truncate(24*time.Hour)) && time.Now().In(location).Hour() >= 20 {
			telegramBot.AnswerCallbackQuery(tgbotapi.NewCallback(callback.ID, "زمان تغییر برای این روز به پایان رسیده است"))
			return
		}

		db := database.Connection().Conn

		var reserve model.Reserve

		query := db.Model(&model.Reserve{})
		query.Where("date = ? ", date)
		query.Where("user_id = ?", user.ID)
		err = query.First(&reserve).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			log.Println(err)
			return
		}

		if err == gorm.ErrRecordNotFound {

			reserve = model.Reserve{
				Date:      date,
				UserID:    user.ID,
				HasLunch:  isLunch,
				HasDinner: isDinner,
			}
		}

		if selectedOption[len(selectedOption)-5:] == "lunch" {

			reserve.HasLunch = !reserve.HasLunch
		} else if selectedOption[len(selectedOption)-6:] == "dinner" {

			reserve.HasDinner = !reserve.HasDinner
		} else {

			log.Println("Invalid option " + selectedOption)
			return
		}

		db.Save(&reserve)

		// Send the updated message to the user
		telegramBot.AnswerCallbackQuery(tgbotapi.NewCallback(callback.ID, fmt.Sprintf("%s تغییر کرد", utils.GetFaDayName(date.Weekday()))))

	}

	// replace showMealSelectionForm with last showMealSelectionForm
	showMealSelectionForm(user, callback.Message.Chat.ID)

	// remove last meal selection message
	telegramBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID:    callback.Message.Chat.ID,
		MessageID: callback.Message.MessageID,
	})

}

// Get the button text depending on whether the meal is selected or not
func getButtonText(meal string, selected bool) string {
	if selected {
		return "✅ " + meal
	}

	return meal
}
