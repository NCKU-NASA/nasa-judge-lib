package user

import (
    "fmt"
    "time"
    "regexp"
    "strings"
    "crypto/rand"
    "crypto/sha256"
    "crypto/hmac"
    "encoding/base64"

    "golang.org/x/exp/slices"

    "github.com/NCKU-NASA/nasa-judge-lib/utils/password"
    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
)

type User struct {
    ID uint `gorm:"primaryKey" json:"-"`
    Username string `gorm:"unique" json:"username"`
    Password password.Password `json:"password"`
    StudentId string `json:"studentId"`
    Email string `gorm:"unique" json:"email"`
    Groups []*Group `gorm:"many2many:user_groups" json:"groups"`
}

func init() {
    database.GetDB().AutoMigrate(&User{})
}

func (user *User) Fix() {
    reg := regexp.MustCompile(`[^0-9a-zA-Z]`)
    user.Username = strings.ToLower(reg.ReplaceAllString(user.Username, ""))
    user.StudentId = strings.ToLower(reg.ReplaceAllString(user.StudentId, ""))
    reg = regexp.MustCompile(`[^0-9a-zA-Z@.]`)
    user.Email = strings.ToLower(reg.ReplaceAllString(user.Email, ""))
}

func GetUser(query any, args ...any) (user User, err error) {
    result := database.GetDB().Model(&User{}).Preload("Groups").Where(query, args...).First(&user)
    err = result.Error
    return
}

func GetUsers() ([]User, error) {
    var resultusers []User
    result := database.GetDB().Model(&User{}).Preload("Groups").Find(&resultusers)
    return resultusers, result.Error
}

func (user *User) Create() error {
    user.Fix()
    if user.Username == "" || user.Password == "" || user.Email == "" {
        return fmt.Errorf("Username, Password and Email can't be empty.")
    }
    result := database.GetDB().Model(&User{}).Preload("Groups").Create(user)
    return result.Error
}

func (user *User) Update() error {
    user.Fix()
    if user.Username == "" || user.Password == "" || user.Email == "" {
        return fmt.Errorf("Username, Password and Email can't be empty.")
    }
    database.GetDB().Model(&User{}).Preload("Groups").Where("username = ?", user.Username).Updates(user)
    return nil
}

func (user User) ContainGroup(group string) bool {
    return slices.ContainsFunc(user.Groups, func(g *Group) bool {
        return g.Groupname == group
    })    
}

func (user User) GenToken(data string) (string, error) {
    salt := make([]byte, 18)
    _, err := rand.Read(salt)
    if err != nil {
        return "", err
    }
    h := hmac.New(sha256.New, []byte(password.GetSecret()))
    h.Write(salt)
    h.Write([]byte(user.Username))
    h.Write([]byte(user.Password))
    h.Write([]byte(data))
    h.Write([]byte(time.Now().Format("2006-01-02T15")))
    return fmt.Sprintf("%s:%s", base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(h.Sum(nil))), nil
}