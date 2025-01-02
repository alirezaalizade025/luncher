package telegramBot

import (
	"encoding/json"
	"fmt"
	"log"
	"luncher/handler/database"
	model "luncher/handler/models"
	"luncher/handler/utils"

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

	bot.Debug = true
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
			}

			user := findUser(db, update)

			if user.ID == 0 {
				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در ارتباط با دیتابیس"))
				continue
			}

			if update.Message.Text == "/select" {

				showMealSelectionForm(user.Reserves, update.Message.Chat.ID)
			}

		}

		// Handle button presses (callback queries)
		if update.CallbackQuery != nil {

			user := findUser(db, update)

			if user.ID == 0 {
				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در ارتباط با دیتابیس"))
				continue
			}

			handleButtonPress(user.Reserves, update.CallbackQuery)
		}
	}
}

func findUser(db *gorm.DB, update tgbotapi.Update) model.User {
	user := model.User{}

	err := db.Model(&model.User{}).Where("telegram_id = ?", update.Message.Chat.ID).First(&user)
	if err.Error != nil {
		log.Println(err.Error)
	}

	return user
}

// Show the meal selection form with inline buttons
func showMealSelectionForm(reserves []model.Reserve, chatID int64) {
	// Get the current meal preferences for the user
	mealPreference := userResponses[chatID]

	// find week number of year
	weekNumber := utils.GetWeekNumber() % 2

	var mealsString string
	meals := [14]string{}

	switch weekNumber {
	case 1:
		mealsString = utils.Getenv("FIRST_WEEK", "")
	case 2:
		mealsString = utils.Getenv("SECOND_WEEK", "")
	default:
		mealsString = utils.Getenv("FIRST_WEEK", "")
	}



	err := json.Unmarshal([]byte(mealsString), &meals)
	if err != nil {
		log.Println(err)
	}


	// Create inline keyboard buttons for each day and meal (lunch and dinner)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("شام", "week_dinner"),
			tgbotapi.NewInlineKeyboardButtonData("نهار", "week_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("*", "nothing"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(meals[0], mealPreference.SaturdayDinner), "saturday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(meals[1], mealPreference.SaturdayLunch), "saturday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("شنبه", "شنبه"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو بی ماهی", mealPreference.SundayDinner), "sunday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.SundayLunch), "sunday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("یکشنبه", "یکشنبه"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو بی ماهی", mealPreference.MondayDinner), "monday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.MondayLunch), "monday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("دوشنبه", "دوشنبه"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو بی ماهی", mealPreference.TuesdayDinner), "tuesday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.TuesdayLunch), "tuesday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("سه شنبه", "سه شنبه"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو بی ماهی", mealPreference.WednesdayDinner), "wednesday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.WednesdayLunch), "wednesday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("چهارشنبه", "چهارشنبه"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو بی ماهی", mealPreference.ThursdayDinner), "thursday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.ThursdayLunch), "thursday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("پنجشنبه", "پنجشنبه"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو بی ماهی", mealPreference.FridayDinner), "friday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.FridayLunch), "friday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData("جمعه", "جمعه"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Please select your meal preferences for each day.")
	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = true
	telegramBot.Send(msg)
}

// Structure to store the meal preferences for each day
type MealPreferences struct {
	MondayLunch     bool
	MondayDinner    bool
	TuesdayLunch    bool
	TuesdayDinner   bool
	WednesdayLunch  bool
	WednesdayDinner bool
	ThursdayLunch   bool
	ThursdayDinner  bool
	FridayLunch     bool
	FridayDinner    bool
	SaturdayLunch   bool
	SaturdayDinner  bool
	SundayLunch     bool
	SundayDinner    bool
}

// Global variable to store user responses
var userResponses = make(map[int64]MealPreferences)

// Handle button press events
func handleButtonPress(reserves []model.Reserve, callback *tgbotapi.CallbackQuery) {
	// Get the user ID and the selected meal option
	userID := int64(callback.From.ID)
	selectedOption := callback.Data

	// Toggle the meal selection for the user
	mealPreference := userResponses[userID]

	var message string

	switch selectedOption {
	case "monday_lunch":
		mealPreference.MondayLunch = !mealPreference.MondayLunch
		message = toggleMessage("Monday", mealPreference.MondayLunch)
	case "monday_dinner":
		mealPreference.MondayDinner = !mealPreference.MondayDinner
		message = toggleMessage("Monday", mealPreference.MondayDinner)
	case "tuesday_lunch":
		mealPreference.TuesdayLunch = !mealPreference.TuesdayLunch
		message = toggleMessage("Tuesday", mealPreference.TuesdayLunch)
	case "tuesday_dinner":
		mealPreference.TuesdayDinner = !mealPreference.TuesdayDinner
		message = toggleMessage("Tuesday", mealPreference.TuesdayDinner)
	case "wednesday_lunch":
		mealPreference.WednesdayLunch = !mealPreference.WednesdayLunch
		message = toggleMessage("Wednesday", mealPreference.WednesdayLunch)
	case "wednesday_dinner":
		mealPreference.WednesdayDinner = !mealPreference.WednesdayDinner
		message = toggleMessage("Wednesday", mealPreference.WednesdayDinner)
	case "thursday_lunch":
		mealPreference.ThursdayLunch = !mealPreference.ThursdayLunch
		message = toggleMessage("Thursday", mealPreference.ThursdayLunch)
	case "thursday_dinner":
		mealPreference.ThursdayDinner = !mealPreference.ThursdayDinner
		message = toggleMessage("Thursday", mealPreference.ThursdayDinner)
	case "friday_lunch":
		mealPreference.FridayLunch = !mealPreference.FridayLunch
		message = toggleMessage("Friday", mealPreference.FridayLunch)
	case "friday_dinner":
		mealPreference.FridayDinner = !mealPreference.FridayDinner
		message = toggleMessage("Friday", mealPreference.FridayDinner)
	case "saturday_lunch":
		mealPreference.SaturdayLunch = !mealPreference.SaturdayLunch
		message = toggleMessage("Saturday", mealPreference.SaturdayLunch)
	case "saturday_dinner":
		mealPreference.SaturdayDinner = !mealPreference.SaturdayDinner
		message = toggleMessage("Saturday", mealPreference.SaturdayDinner)
	case "sunday_lunch":
		mealPreference.SundayLunch = !mealPreference.SundayLunch
		message = toggleMessage("Sunday", mealPreference.SundayLunch)
	case "sunday_dinner":
		mealPreference.SundayDinner = !mealPreference.SundayDinner
		message = toggleMessage("Sunday", mealPreference.SundayDinner)
	case "week_lunch":
		mealPreference.MondayLunch = true
		mealPreference.TuesdayLunch = true
		mealPreference.WednesdayLunch = true
		mealPreference.ThursdayLunch = true
		mealPreference.FridayLunch = true
		mealPreference.SaturdayLunch = true
		mealPreference.SundayLunch = true
		message = "All lunches selected"
	case "week_dinner":
		mealPreference.MondayDinner = true
		mealPreference.TuesdayDinner = true
		mealPreference.WednesdayDinner = true
		mealPreference.ThursdayDinner = true
		mealPreference.FridayDinner = true
		mealPreference.SaturdayDinner = true
		mealPreference.SundayDinner = true
		message = "All dinners selected"
	case "nothing":
		mealPreference.MondayLunch = true
		mealPreference.TuesdayLunch = true
		mealPreference.WednesdayLunch = true
		mealPreference.ThursdayLunch = true
		mealPreference.FridayLunch = true
		mealPreference.SaturdayLunch = true
		mealPreference.SundayLunch = true
		mealPreference.MondayDinner = true
		mealPreference.TuesdayDinner = true
		mealPreference.WednesdayDinner = true
		mealPreference.ThursdayDinner = true
		mealPreference.FridayDinner = true
		mealPreference.SaturdayDinner = true
		mealPreference.SundayDinner = true
		message = "All meals selected"
	}

	userResponses[userID] = mealPreference

	// Send the updated message to the user
	telegramBot.AnswerCallbackQuery(tgbotapi.NewCallback(callback.ID, message))

	// replace showMealSelectionForm with last showMealSelectionForm
	showMealSelectionForm(reserves, callback.Message.Chat.ID)

	// remove last meal selection message
	telegramBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID:    callback.Message.Chat.ID,
		MessageID: callback.Message.MessageID,
	})

}

// Toggle the message to show if the meal is selected or not
func toggleMessage(meal string, selected bool) string {

	if selected {
		return fmt.Sprintf("%s: Selected", meal)
	}
	return fmt.Sprintf("%s: Not Selected", meal)
}

// Get the button text depending on whether the meal is selected or not
func getButtonText(meal string, selected bool) string {
	if selected {
		return "✅ " + meal
	}

	return meal
}
