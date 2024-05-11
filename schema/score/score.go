package score

import (
    "time"

    "gorm.io/datatypes"

    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
)

type Score struct {
    ID uint `gorm:"primaryKey" json:"-"`
    UserID uint `json:"-"`
    User user.User `gorm:"foreignKey:UserID" json:"user"`
    LabId string `json:"labId"`
    Score int `json:"score"`
    Result datatypes.JSON `json:"result"`
    Date datatypes.JSON `json:"data"`
    CreatedAt time.Time `json:"createAt"`
}

func init() {
    database.GetDB().AutoMigrate(&Score{})
}

