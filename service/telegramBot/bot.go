package telegramBot

import (
	"fmt"
	"log"
	"luncher/handler/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	// Loop to listen for incoming messages or button presses
	for update := range updates {
		if update.Message != nil {
			// Start command to show the meal selection form
			if update.Message.Text == "/start" {
				showMealSelectionForm(telegramBot, update.Message.Chat.ID)
			}
		}

		// Handle button presses (callback queries)
		if update.CallbackQuery != nil {
			handleButtonPress(telegramBot, update.CallbackQuery)
		}
	}
}

// Show the meal selection form with inline buttons
func showMealSelectionForm(bot *tgbotapi.BotAPI, chatID int64) {
	// Get the current meal preferences for the user
	mealPreference := userResponses[chatID]

	// Create inline keyboard buttons for each day and meal (lunch and dinner)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("نهار", "نهار"),
			tgbotapi.NewInlineKeyboardButtonData("شام", "شام"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Monday", mealPreference.MondayLunch), "monday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Monday", mealPreference.MondayDinner), "monday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Monday", mealPreference.MondayDinner), "monday_dinner"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Tuesday", mealPreference.TuesdayLunch), "tuesday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Tuesday", mealPreference.TuesdayDinner), "tuesday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Tuesday", mealPreference.TuesdayDinner), "tuesday_dinner"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Wednesday", mealPreference.WednesdayLunch), "wednesday_lunch"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("سبزی پلو با ماهی", mealPreference.WednesdayDinner), "wednesday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Wednesday", mealPreference.TuesdayDinner), "wednesday_dinner2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Thursday", mealPreference.ThursdayLunch), "thursday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Thursday", mealPreference.ThursdayDinner), "thursday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Thursday", mealPreference.ThursdayDinner), "thursday_dinner"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Friday", mealPreference.FridayLunch), "friday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Friday", mealPreference.FridayDinner), "friday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Friday", mealPreference.FridayDinner), "friday_dinner"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Saturday", mealPreference.SaturdayLunch), "saturday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Saturday", mealPreference.SaturdayDinner), "saturday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Saturday", mealPreference.SaturdayDinner), "saturday_dinner"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Sunday", mealPreference.SundayLunch), "sunday_lunch"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Sunday", mealPreference.SundayDinner), "sunday_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("Sunday", mealPreference.SundayDinner), "sunday_dinner"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Please select your meal preferences for each day.")
	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = true
	bot.Send(msg)
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
func handleButtonPress(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
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
	}

	userResponses[userID] = mealPreference

	// Send the updated message to the user
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(callback.ID, message))

	// remove last meal selection message
	bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID:    callback.Message.Chat.ID,
		MessageID: callback.Message.MessageID,
	})

	// Update the meal selection form with the current selection
	showMealSelectionForm(bot, callback.Message.Chat.ID)
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
