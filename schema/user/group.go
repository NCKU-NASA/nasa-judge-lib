package user

import (
    "encoding/json"
    "gopkg.in/yaml.v3"
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

func (c Group) MarshalYAML() (any, error) {
    return c.Groupname, nil
}

func (c *Group) UnmarshalYAML(b *yaml.Node) error {
    var tmp string
    err := b.Decode(&tmp)
    *c = Group{
        Groupname: tmp,
    }
    return err
}
