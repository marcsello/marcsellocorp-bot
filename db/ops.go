package db

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"time"
)

func GetUserById(id int64) (*User, error) {
	// preload subs
	var user User
	result := db.Preload("Subscriptions").Take(&user, id)

	if result.Error == nil && result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &user, result.Error
}

func GetAllChannels() ([]Channel, error) {
	var channels []Channel
	result := db.Find(&channels)
	return channels, result.Error
}

func GetChannelByName(name string) (*Channel, error) {
	var channel Channel
	result := db.Preload("Subscribers").Where("name = ?", name).First(&channel)

	if result.Error == nil && result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &channel, result.Error
}

func GetChannelById(id uint) (*Channel, error) {
	var channel Channel
	result := db.Preload("Subscribers").Take(&channel, id)

	if result.Error == nil && result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &channel, result.Error
}

func isPgError(err error, severity, code string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Severity == severity && pgErr.Code == code
	}
	return false
}

// ChangeSubscription adds or deletes a subscription, the first return value indicates if anything changed or not
func ChangeSubscription(userId int64, channelId uint, subscribed bool) (bool, error) {
	var result *gorm.DB
	if subscribed {

		result = db.Exec("INSERT INTO subscriptions (user_id, channel_id) VALUES (?, ?)", userId, channelId)

		if result.Error != nil {
			if isPgError(result.Error, "ERROR", "23505") { // duplicate key
				return false, nil
			}
			return false, result.Error
		}
		return true, nil

	} else {

		result = db.Exec("DELETE FROM subscriptions WHERE user_id = ? AND channel_id = ?", userId, channelId)

		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil

	}
}

func GetAndUpdateTokenByHash(tokenHashBytes []byte) (*Token, error) {
	var token Token
	var found bool
	err := db.Transaction(func(tx *gorm.DB) error {
		// fetch allowed ch as well

		result := tx.Preload("Subscribers").Where("token_hash = ?", tokenHashBytes).First(&token)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		found = true

		result = tx.Model(&token).Update("last_used", time.Now())
		return result.Error
	})
	if err != nil {
		return nil, err
	}
	if found {
		return &token, nil
	} else {
		return nil, nil
	}
}
