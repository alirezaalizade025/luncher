package telegramBot

import (
	"encoding/json"
	"fmt"
	"log"
	"luncher/handler/database"
	model "luncher/handler/models"
	"luncher/handler/utils"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gorm.io/gorm"
)

var telegramBot *tgbotapi.BotAPI
var memCache *utils.Store

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

	// setup memCache
	memCache = utils.MemCache()
	go memCache.Cleanup()

	// Loop to listen for incoming messages or button presses
	for update := range updates {

		if update.Message != nil {

			if _, found := memCache.Get(fmt.Sprintf("%s_set_meal", update.Message.From.UserName)); found {

				handleSetMealName(update, db)
				continue
			}

			if update.Message.Text == "/setList" {
				if !isAdmin(update.Message.From.UserName) {

					telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "شما دسترسی ندارید."))

					log.Printf("Unauthorized access - username: %s", update.Message.From.UserName)
					continue
				} else {

					showMealSetFrom(update.Message.Chat.ID)
				}

				continue
			}

			if update.Message.Text == "/getCounts" {
				showCounts(update, db)

				continue
			}

			if update.Message.Text == "/getReserves" {

				showReservesDetails(update, db)
				continue
			}

			// Start command to show the meal selection form
			if update.Message.Text == "/start" {

				createUser(update, db)

				helpStr := helpMessageCreator(update)

				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, helpStr.String()))
			}

			user := findUser(db, update.Message.Chat.ID)

			if user.ID == 0 {

				createUser(update, db)
			}

			if update.Message.Text == "/select" {

				showMealSelectionForm(user, update.Message.Chat.ID)
			}

			if update.Message.Text == "/help" {

				helpStr := helpMessageCreator(update)

				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, helpStr.String()))
			}

			if update.Message.Text == "/setting" {

				showSettingForm(user, update.Message.Chat.ID)
			}

		}

		// Handle button presses (callback queries)
		if update.CallbackQuery != nil {

			if strings.HasPrefix(update.CallbackQuery.Data, "set_lunch_") ||
				strings.HasPrefix(update.CallbackQuery.Data, "set_dinner_") {

				handleSetMealList(update)
				continue
			}

			//find user id
			user := findUser(db, int64(update.CallbackQuery.From.ID))

			if user.ID == 0 {
				telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در ارتباط با دیتابیس"))
				continue
			}

			if strings.HasPrefix(update.CallbackQuery.Data, "setting_") {

				if update.CallbackQuery.Data == "setting_always_lunch" {

					user.AlwaysLunch = !user.AlwaysLunch
					db.Save(&user)
				}

				if update.CallbackQuery.Data == "setting_always_dinner" {

					user.AlwaysDinner = !user.AlwaysDinner
					db.Save(&user)
				}

				telegramBot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "تغییر کرد"))

				showSettingForm(user, int64(update.CallbackQuery.From.ID))

				// remove last setting form
				_, err := telegramBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
					ChatID:    int64(update.CallbackQuery.From.ID),
					MessageID: update.CallbackQuery.Message.MessageID,
				})

				if err != nil {
					log.Println(err)
				}

				continue
			}

			handleButtonPress(user, update.CallbackQuery)
		}
	}
}

func helpMessageCreator(update tgbotapi.Update) strings.Builder {
	helpStr := strings.Builder{}

	helpStr.WriteString("راهنما:\n")
	helpStr.WriteString("/select - انتخاب غذا\n")
	helpStr.WriteString("\t\t\tوعده های غذایی دو هفته‌ی آینده نمایش داده میشود و قابل اضافه و حذف شدن هستند. ( توجه داشته باشید که وعده هر روز نهایتا تا ساعت ۲۰ روز قبل، قابل تغییر میباشد)\n")
	helpStr.WriteString("/setting - تنظیمات\n")
	helpStr.WriteString("\t\t\tاگر گزینه همیشه نهار یا همیشه شام را انتخاب کنید به صورت کامل تمام آن وعده در کل هفت روز هفته انتخاب شده و نیاز به انخاب دانه دانه نمیباشد.\n")

	if isAdmin(update.Message.From.UserName) {

		helpStr.WriteString("\n\n")
		helpStr.WriteString("تنظیمات مخصوص ادمین:\n")
		helpStr.WriteString("/setList - ویرایش لیست غذا دو هفته ای\n")
		helpStr.WriteString("/getCounts - نمایش تعداد امروز\n")
		helpStr.WriteString("/getReserves - نمایش جزئیات دو هفته آینده\n")
	}
	return helpStr
}

func createUser(update tgbotapi.Update, db *gorm.DB) model.User {
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

	return user
}

func showCounts(update tgbotapi.Update, db *gorm.DB) {
	if !isAdmin(update.Message.From.UserName) {

		telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "شما دسترسی ندارید."))

		log.Printf("Unauthorized access - username: %s", update.Message.From.UserName)
		return
	}

	var alwaysLunchCounts int64
	var todayHasLunchCounts int64
	db.Model(&model.User{}).Where("always_lunch = ?", true).Count(&alwaysLunchCounts)
	query := db.Model(&model.Reserve{})
	query = query.Where("date = ? AND has_lunch = ?", time.Now().Truncate(24*time.Hour), true)
	query = query.Where("(SELECT COUNT(*) FROM users WHERE users.id = reserves.user_id AND users.always_lunch = true) = 0")
	query.Count(&todayHasLunchCounts)

	var alwaysDinnerCounts int64
	var todayHasDinnerCounts int64
	db.Model(&model.User{}).Where("always_dinner = ?", true).Count(&alwaysDinnerCounts)
	query2 := db.Model(&model.Reserve{})
	query2 = query2.Where("date = ? AND has_dinner = ?", time.Now().Truncate(24*time.Hour), true)
	query2 = query2.Where("(SELECT COUNT(*) FROM users WHERE users.id = reserves.user_id AND users.always_dinner = true) = 0")
	query2.Count(&todayHasDinnerCounts)

	statsMessage := fmt.Sprintf("نهار: %d\nشام: %d", todayHasLunchCounts+alwaysLunchCounts, todayHasDinnerCounts+alwaysDinnerCounts)

	telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, statsMessage))
}

func showReservesDetails(update tgbotapi.Update, db *gorm.DB) {
	if !isAdmin(update.Message.From.UserName) {

		telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "شما دسترسی ندارید."))

		log.Printf("Unauthorized access - username: %s", update.Message.From.UserName)
		return
	}

	var statsMessage strings.Builder
	today := time.Now()

	for i := 0; i < 14; i++ {
		var lunchUsers []model.User

		date := today.AddDate(0, 0, i).Truncate(24 * time.Hour).Format("2006-01-02")

		db.Model(&model.User{}).
			Where("always_lunch = ? OR id IN (SELECT user_id FROM reserves WHERE has_lunch = ? AND date = ?)", true, true, date).
			Find(&lunchUsers)

		var dinnerUsers []model.User
		db.Model(&model.User{}).
			Where("always_dinner = ? OR id IN (SELECT user_id FROM reserves WHERE has_dinner = ? AND date = ?)", true, true, date).
			Find(&dinnerUsers)

		lunchUsernames := []string{}
		for _, user := range lunchUsers {
			lunchUsernames = append(lunchUsernames, fmt.Sprintf("@%s", user.Username))
		}

		dinnerUsernames := []string{}
		for _, user := range dinnerUsers {
			dinnerUsernames = append(dinnerUsernames, fmt.Sprintf("@%s", user.Username))
		}

		Date := today.AddDate(0, 0, i)

		jalaliDateYear, jalaliDateMonth, jalaliDateDay := utils.GregorianToJalali(Date.Year(), int(Date.Month()), Date.Day())

		statsMessage.WriteString(fmt.Sprintf(
			"%s\n\nlunch: %d\n%s\n\ndinner: %d\n%s\n\n----------\n",

		fmt.Sprintf("%d/%d/%d", jalaliDateYear, jalaliDateMonth, jalaliDateDay),
			len(lunchUsers),
			strings.Join(lunchUsernames, "\n"),
			len(dinnerUsers),
			strings.Join(dinnerUsernames, "\n"),
		))
	}

	telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, statsMessage.String()))
}

func handleSetMealName(update tgbotapi.Update, db *gorm.DB) {
	mealData, _ := memCache.Get(fmt.Sprintf("%s_set_meal", update.Message.From.UserName))
	mealIDString := mealData.(map[string]string)["mealID"]
	mealType := mealData.(map[string]string)["mealType"]

	mealID, _ := strconv.Atoi(mealIDString)

	meal := model.Meal{
		ID: uint(mealID),
	}

	db.First(&meal)

	if mealType == "lunch" {
		meal.Lunch = &update.Message.Text
	} else if mealType == "dinner" {
		meal.Dinner = &update.Message.Text
	}

	db.Save(&meal)

	telegramBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "بروزرسانی شد."))
	memCache.Delete(fmt.Sprintf("%s_set_meal", update.Message.From.UserName))

	showMealSetFrom(int64(update.Message.Chat.ID))
}

func handleSetMealList(update tgbotapi.Update) {
	if !isAdmin(update.CallbackQuery.From.UserName) {

		telegramBot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "شما دسترسی ندارید."))

		log.Printf("Unauthorized access - username: %s", update.CallbackQuery.From.UserName)
		return
	}

	mealID := strings.Split(update.CallbackQuery.Data, "_")[2]
	mealID = strings.Trim(mealID, " ")

	mealType := strings.Split(update.CallbackQuery.Data, "_")[1]
	mealType = strings.Trim(mealType, " ")

	memCacheData := map[string]string{
		"mealID":   mealID,
		"mealType": mealType,
	}

	memCache.Set(fmt.Sprintf("%s_set_meal", update.CallbackQuery.From.UserName), memCacheData, 30*time.Second)

	telegramBot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Enter %s for day %s:", mealType, mealID)))
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

	lunchesList, dinnersList := generateMeals()

	for i := 0; i < 14; i++ {

		date := time.Now().AddDate(0, 0, i)

		// find week number of year
		weekNumber := utils.GetJalaliWeekNumber(date)
		weekDay := date.Weekday()
		faDayNumber := utils.GetJalaliWeekDayNumber(weekDay)
		faDayName := utils.GetFaDayName(weekDay)
		_, faMonth, faDay := utils.GregorianToJalali(date.Year(), int(date.Month()), date.Day())

		index := int(weekNumber*7+int(faDayNumber)) - 1

		// Check if the user has already selected a meal for this day
		selectedMeal := model.Reserve{
			HasDinner: user.AlwaysDinner,
			HasLunch:  user.AlwaysLunch,
		}

		for _, reserve := range next14DaysReserves {

			if reserve.Date.Format("2006-01-02") == date.Format("2006-01-02") {

				selectedMeal.HasDinner = reserve.HasDinner
				selectedMeal.HasLunch = reserve.HasLunch
				break
			}
		}

		key := fmt.Sprintf("%s (%d/%d)", faDayName, faMonth, faDay)

		rowButton := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(dinnersList[index], selectedMeal.HasDinner), fmt.Sprintf("%s_dinner", date.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(lunchesList[index], selectedMeal.HasLunch), fmt.Sprintf("%s_lunch", date.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(key, date.Format("2006-01-02")),
		)

		buttons = append(buttons, rowButton)
	}

	// Create inline keyboard buttons for each day and meal (lunch and dinner)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "Please select your meal preferences for each day.")
	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = true
	message, err := telegramBot.Send(msg)
	if err != nil {
		log.Println("show meal error", err)
		return
	}

	memCacheKey := fmt.Sprintf("user_%d_last_message", chatID)
	memCache.Set(memCacheKey, message.MessageID, 1*time.Minute)
}

func showMealSetFrom(chatID int64) {

	buttons := [][]tgbotapi.InlineKeyboardButton{}

	var next14DaysMeals []model.Meal
	db := database.Connection().Conn
	db.Order("id").Find(&next14DaysMeals)

	for i := 1; i < 15; i++ {

		// find day record by when next14DaysMeals.id = i
		var dayMeal model.Meal
		for _, meal := range next14DaysMeals {
			if meal.ID == uint(i) {
				dayMeal = meal
				break
			}
		}

		if dayMeal.Dinner == nil {
			dinner := "شام"
			dayMeal.Dinner = &dinner
		}

		if dayMeal.Lunch == nil {
			lunch := "نهار"
			dayMeal.Lunch = &lunch
		}

		weekNumber := i
		if i > 7 {
			weekNumber = i - 7
		}

		rowButton := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(*dayMeal.Dinner, fmt.Sprintf("set_dinner_%d", i)),
			tgbotapi.NewInlineKeyboardButtonData(*dayMeal.Lunch, fmt.Sprintf("set_lunch_%d", i)),
			tgbotapi.NewInlineKeyboardButtonData(utils.GetFaDayNameByNumber(weekNumber), "d"),
		)

		buttons = append(buttons, rowButton)
	}

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "انتخاب کنید.")
	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = true
	_, err := telegramBot.Send(msg)
	if err != nil {
		log.Println("show meal list error", err)
		return
	}
}

func showSettingForm(user model.User, chatID int64) {

	buttons := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("همیشه شام", user.AlwaysDinner), "setting_always_dinner"),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText("همیشه نهار", user.AlwaysLunch), "setting_always_lunch"),
		),
	}

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "تنظیمات کلی")
	msg.ReplyMarkup = inlineKeyboard
	msg.DisableNotification = true
	_, err := telegramBot.Send(msg)
	if err != nil {
		log.Println("show setting error", err)
		return
	}
}

func generateMeals() ([14]string, [14]string) {

	db := database.Connection().Conn

	var meals []model.Meal
	db.Model(&model.Meal{}).Find(&meals)

	lunches := [14]string{}
	dinners := [14]string{}

	// fill empty meals
	for i := 0; i < 14; i++ {

		// check if meal with id = i exists place it
		for _, m := range meals {

			if m.ID == uint(i+1) {

				if m.Lunch == nil {
					lunches[i] = "نهار"
				} else {
					lunches[i] = *m.Lunch
				}

				if m.Dinner == nil {
					dinners[i] = "شام"
				} else {
					dinners[i] = *m.Dinner
				}

				break
			}
		}

		if lunches[i] == "" {
			lunches[i] = "نهار"
		}

		if dinners[i] == "" {
			dinners[i] = "شام"
		}
	}

	return lunches, dinners
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

			// max edit time is yesterday 17:30
			maxEditTime := date.AddDate(0, 0, -1).Truncate(24 * time.Hour).Add(17*time.Hour + 30*time.Minute)

			// if date is today and hour pass from 20 in tehran
			// show error message
			if time.Now().After(maxEditTime) {
				continue
			}

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

		// max edit time is yesterday 17:30
		maxEditTime := date.AddDate(0, 0, -1).Truncate(24 * time.Hour).Add(17*time.Hour + 30*time.Minute)

		// if date is today and hour pass from 20 in tehran
		// show error message
		if time.Now().After(maxEditTime) {
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

	var lastMessageID int
	item, found := memCache.Get(fmt.Sprintf("user_%d_last_message", callback.From.ID))
	if !found {
		lastMessageID = callback.Message.MessageID
	} else {
		lastMessageID = item.(int)
	}

	// remove last meal selection message
	_, err := telegramBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID:    callback.Message.Chat.ID,
		MessageID: lastMessageID,
	})

	if err != nil {
		log.Println(err)
	}

	// replace showMealSelectionForm with last showMealSelectionForm
	showMealSelectionForm(user, callback.Message.Chat.ID)

}

// Get the button text depending on whether the meal is selected or not
func getButtonText(meal string, selected bool) string {
	if selected {
		return "✅ " + meal
	}

	return meal
}

func isAdmin(username string) bool {
	admins := utils.Getenv("ADMINS", "")
	var adminList []string
	err := json.Unmarshal([]byte(admins), &adminList)
	if err != nil {
		log.Println(err)
		return false
	}

	for _, admin := range adminList {
		if admin == username {
			return true
		}
	}

	return false
}
