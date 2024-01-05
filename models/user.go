package models

import "time"

type User struct {
	Id        uint64    `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	ApiKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "user"
}

func GetLoginUser(username string) (u User, err error) {
	u = User{}
	if err = DB().Model(&u).Where("email", username).Take(&u).Error; err != nil {
		return
	}
	return
}
