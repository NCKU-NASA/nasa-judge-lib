package user

import (
    "encoding/json"
)

type Group struct {
    ID uint `gorm:"primaryKey" json:"-"`
    Groupname string `gorm:"unique" json:"groupname"`
    Members []*User `gorm:"many2many:user_groups"`
}

func (c Group) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.Groupname)
}

func (c *Group) UnmarshalJSON(b []byte) error {
    *c = Group{}
    return json.Unmarshal(b, &(c.Groupname))
}

