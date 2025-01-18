package utils

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func LoadENV() {
	err := godotenv.Load()

	if err != nil {
		if os.IsNotExist(err) {

			log.Println(".env file does not exist, using default values")
		} else {
			log.Fatalln("Error loading .env file:", err)
		}
	}
}

func GetWeekNumber(date time.Time) int {

	_, week := date.ISOWeek()

	return week
}

func GetFaDayName(weekDay time.Weekday) string {
	switch weekDay {
	case time.Saturday:
		return "شنبه"
	case time.Sunday:
		return "یکشنبه"
	case time.Monday:
		return "دوشنبه"
	case time.Tuesday:
		return "سه شنبه"
	case time.Wednesday:
		return "چهارشنبه"
	case time.Thursday:
		return "پنجشنبه"
	case time.Friday:
		return "جمعه"
	default:
		return ""
	}
}

func GetJalaliWeekDayNumber(weekDay time.Weekday) int {
	switch weekDay {
	case time.Saturday:
		return 1
	case time.Sunday:
		return 2
	case time.Monday:
		return 3
	case time.Tuesday:
		return 4
	case time.Wednesday:
		return 5
	case time.Thursday:
		return 6
	case time.Friday:
		return 7
	default:
		return 0
	}
}

func GetJalaliDate(date time.Time) string {
	return ""
}

func GregorianToJalali(year, month, day int) (int, int, int) {
	// Convert Gregorian date to Julian Day Number (JDN)
	if month <= 2 {
		year--
		month += 12
	}
	a := int(year / 100)
	b := 2 - a + int(a/4)
	jdn := int(365.25*float64(year+4716)) + int(30.6001*float64(month+1)) + day + b - 1524

	// Jalali Epoch Offset
	jalaliEpochOffset := 1948320.5
	daysSinceJalaliEpoch := float64(jdn) - jalaliEpochOffset

	// Calculate Jalali year
	jalaliYear := int(daysSinceJalaliEpoch / 365.24219858156)
	remainingDays := daysSinceJalaliEpoch - (float64(jalaliYear) * 365.24219858156)

	// Adjust for leap years
	if remainingDays < 0 {
		jalaliYear--
		remainingDays += 365.24219858156
	}

	// Determine Jalali month and day
	var jalaliMonth, jalaliDay int
	var monthDays = []int{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29} // Jalali month lengths

	for i, days := range monthDays {
		if remainingDays < float64(days) {
			jalaliMonth = i + 1
			jalaliDay = int(math.Floor(remainingDays)) + 1
			break
		}
		remainingDays -= float64(days)
	}

	return jalaliYear, jalaliMonth, jalaliDay
}
