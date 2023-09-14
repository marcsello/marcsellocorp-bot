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

func GetAllTokens() ([]Token, error) {
	var tokens []Token
	result := db.Preload("AllowedChannels").Omit("token_hash").Find(&tokens)
	return tokens, result.Error
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

func CreateChannel(channel *Channel) (*Channel, error) {
	result := db.Save(channel)

	if result.Error != nil {
		if isPgError(result.Error, "ERROR", "23505") { // duplicate key
			return nil, gorm.ErrDuplicatedKey
		}
		return nil, result.Error
	}
	return channel, nil
}

func CreateToken(token *Token, allowedChannelNames []string) (*Token, error) {
	err := db.Transaction(func(tx *gorm.DB) error {

		var channels []Channel
		result := tx.Where("name IN ?", allowedChannelNames).Find(&channels)
		if result.RowsAffected != int64(len(allowedChannelNames)) {
			return gorm.ErrRecordNotFound
		}

		// turn array to array of pointers... for.. some... reason that's beyond me
		token.AllowedChannels = make([]*Channel, len(channels))
		for i := range channels {
			token.AllowedChannels[i] = &channels[i]
		}

		result = tx.Save(token)
		if result.Error != nil {
			if isPgError(result.Error, "ERROR", "23505") { // duplicate key
				return gorm.ErrDuplicatedKey
			}
			return result.Error
		}
		return nil
	})
	return token, err
}

func DeleteChannelByName(name string) error {
	result := db.Where("name = ?", name).Delete(&Channel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func DeleteTokenByName(name string) error {
	result := db.Unscoped().Where("name = ?", name).Delete(&Token{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

		result := tx.Preload("AllowedChannels").Where("token_hash = ?", tokenHashBytes).First(&token)
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
