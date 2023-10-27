package background

import (
	"database/sql/driver"
	"encoding/json"
)

type User struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"not null"`
	Surname     string `json:"surname" gorm:"not null"`
	Patronymic  string `json:"patronymic"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"country_id"`
}

func (d User) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *User) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &d)
}
