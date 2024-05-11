package password

import (
    "fmt"
    "strings"
    "reflect"
    "crypto/rand"
    "crypto/hmac"
    "crypto/sha512"
    "encoding/base64"
    "encoding/json"
//    "encoding/hex"
    "database/sql/driver"
    "github.com/GehirnInc/crypt"
    _ "github.com/GehirnInc/crypt/sha512_crypt"
)

type Password string

type password struct {
    Type string `json:"type"`
    Password string `json:"password"`
}

var crypter crypt.Crypter
var secret string

func init() {
    crypter = crypt.SHA512.New()
}

func Init(initsecret string) {
    if secret == "" {
        secret = initsecret
    }
}

func GetSecret() string {
    return secret
}

func New(pass string) Password {
    if pass == "" {
        return ""
    }
    salt := make([]byte, 18)
    _, err := rand.Read(salt)
    if err != nil {
        panic("Generate password error!")
    }
    return newshadow(salt, pass)
}

func newhmac(salt []byte, pass string) Password {
    h := hmac.New(sha512.New, []byte(secret))
    h.Write(salt)
    h.Write([]byte(pass))
    return Password(fmt.Sprintf("%s$%s", base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(h.Sum(nil))))
}

func newshadow(salt []byte, pass string) Password {
    ret, err := crypter.Generate([]byte(pass), []byte(fmt.Sprintf("$6$%s", base64.StdEncoding.EncodeToString(salt))))
    if err != nil {
        panic("Generate password error!")
    }
    return Password(ret)
}

func (c Password) Verify(key string) bool {
    return (key == "" && c == "") || c.verifyshadow(key)
}

func (c Password) verifyhmac(key string) bool {
    passpart := strings.Split(string(c), "$")
    salt, err := base64.StdEncoding.DecodeString(passpart[0])
    if err != nil {
        panic("Verify password error!")
    }
    return newhmac(salt, key) == c
}

func (c Password) verifyshadow(key string) bool {
    return crypter.Verify(string(c), []byte(key)) == nil
}

func (c Password) MarshalJSON() ([]byte, error) {
    pass := password{
        Type: "secret",
        Password: string(c),
    }
    return json.Marshal(pass)
}

func (c *Password) UnmarshalJSON(b []byte) error {
    var tmpstr string
    err := json.Unmarshal(b, &tmpstr)
    if err == nil {
        *c = New(tmpstr)
        return err
    }
    var tmppass password
    err = json.Unmarshal(b, &tmppass)
    if err != nil {
        return err
    }
    switch strings.ToLower(tmppass.Type) {
    case "plain":
        *c = New(tmppass.Password)
    case "secret":
        *c = Password(tmppass.Password)
    default:
        return fmt.Errorf("Invalid type %s", strings.ToLower(tmppass.Type))
    }
    return nil
}

func (c *Password) Scan(value interface{}) (err error) {
    if val, ok := value.(string); ok {
        *c = Password(val)
    } else if val, ok := value.([]uint8); ok {
        *c = Password(val)
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c Password) Value() (driver.Value, error) {
    return string(c), nil
}
