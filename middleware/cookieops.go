package middleware

import (
	"errors"
	"time"

	"github.com/gobuffalo/buffalo"
)

func CreateCookie(c buffalo.Context, userID string) {
	c.Cookies().Set("user_id", userID, 200*365*24*time.Hour)
	c.Cookies().Set("date", time.Now().Format(time.RFC3339), 200*365*24*time.Hour)
}

func RemoveCookie(c buffalo.Context) {
	c.Cookies().Delete("user_id")
	c.Cookies().Delete("date")
}

func GetCookieUserID(c buffalo.Context) (userID string, err error) {
	userID, err = c.Cookies().Get("user_id")
	if err != nil {
		return "", err
	} else if userID == "" {
		return "", errors.New("user_id cookie not found")
	}
	return userID, nil
}

func GetCookie(c buffalo.Context) (userID string, date time.Time, err error) {
	userID, err = c.Cookies().Get("user_id")
	if err != nil {
		return "", time.Time{}, err
	} else if userID == "" {
		return "", time.Time{}, errors.New("user_id cookie not found")
	}

	dateStr, err := c.Cookies().Get("date")
	if err != nil {
		return "", time.Time{}, err
	} else if dateStr == "" {
		return "", time.Time{}, errors.New("date cookie not found")
	}

	date, err = time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "", time.Time{}, errors.New("failed to parse date cookie")
	}

	return userID, date, nil
}
