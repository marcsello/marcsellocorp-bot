package db

func GetUserById(id int64) (*User, error) {
	// preload subs
}

func GetAllChannels() ([]Channel, error) {

}

func GetChannelByName(name string) (*Channel, error) {
	// fill subs
}

func GetChannelById(id uint) (*Channel, error) {
	// fill subs
}

func ChangeSubscription(userId int64, channelId uint, subscribed bool) error {

}

func GetAndUpdateTokenByHash(tokenHashBytes []byte) (*Token, error) {
	// use tx
	// fetch allowed ch as well
}

func NewPendingQuestion(q *PendingQuestion) (*PendingQuestion, error) {

}
