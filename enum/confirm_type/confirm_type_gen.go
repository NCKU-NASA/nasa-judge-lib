package confirm_type

import (
    "strings"
    "strconv"
    "context"
    "math/rand"
    "encoding/json"
    "fmt"
    "database/sql/driver"
    "reflect"

    "gorm.io/datatypes"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "gorm.io/gorm/schema"
    "gopkg.in/yaml.v3"
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/config"
    
    
    "golang.org/x/exp/slices"
    
    "text/template"
    
)

type ConfirmType byte
type ConfirmTypes []ConfirmType

const (
    NewAccount ConfirmType = iota
    ForgetPassword
    Unknown ConfirmType = 255
)
var set []string
var subjectset []string
var templateset []*template.Template
func init() {
    set = []string{"NewAccount", "ForgetPassword"}
    subjectset = []string{"Activate NASAJudge account", "Restore NASAJudge account"}
    templateset = []*template.Template{
        template.Must(template.New("newaccount.tmpl").ParseFiles("templates/confirmmail/newaccount.tmpl")),
        template.Must(template.New("forgetpassword.tmpl").ParseFiles("templates/confirmmail/forgetpassword.tmpl")),
    }
}
func (c ConfirmType) GetSubject() string {
    return subjectset[c]
}
func (c ConfirmType) GetTemplate() *template.Template {
    return templateset[c]
}

func (c ConfirmType) String() string {
    if c == Unknown {
        return ""
    }
    return set[c]
}

func FromString(s string) ConfirmType {
    if s == "" {
        return Unknown
    }
    return ConfirmType(byte(slices.Index(set, s)))
}

func Random() ConfirmType {
    return ConfirmType(rand.Intn(Len()))
}

func Len() int {
    return len(set)
}

func (c ConfirmType) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.String())
}

func (c ConfirmType) MarshalYAML() (any, error) {
    return c.String(), nil
}

func (c *ConfirmType) UnmarshalJSON(b []byte) error {
    var num int
    err := json.Unmarshal(b, &num)
    if err == nil {
        *c = ConfirmType(num)
    } else {
        var tmp string
        err = json.Unmarshal(b, &tmp)
        if err != nil {
            return err
        }
        *c = FromString(tmp)
        if tmp != "" && *c == Unknown {
            err = fmt.Errorf("Invalid param %s", tmp)
        }
    }
    return err
}

func (c *ConfirmType) UnmarshalYAML(b *yaml.Node) error {
    var num int
    err := b.Decode(&num)
    if err == nil {
        *c = ConfirmType(num)
    } else {
        var tmp string
        err = b.Decode(&tmp)
        if err != nil {
            return err
        }
        *c = FromString(tmp)
        if b.Value != "" && *c == Unknown {
            err = fmt.Errorf("Invalid param %s", tmp)
        }
    }
    return err
}

func (c ConfirmTypes) MarshalJSON() ([]byte, error) {
    return json.Marshal([]ConfirmType(c))
}

func (c ConfirmTypes) MarshalYAML() (any, error) {
    return []ConfirmType(c), nil
}

func (c *ConfirmTypes) UnmarshalJSON(b []byte) error {
    var tmp []ConfirmType
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = ConfirmTypes(tmp)
    return err
}

func (c *ConfirmTypes) UnmarshalYAML(b *yaml.Node) error {
    var tmp []ConfirmType
    err := b.Decode(&tmp)
    if err != nil {
        return err
    }
    *c = ConfirmTypes(tmp)
    return err
}

func (c *ConfirmTypes) Scan(value interface{}) (err error) {
    switch config.DBservice {
    case "mysql", "sqlite":
        if val, ok := value.(datatypes.JSON); ok {
            err = json.Unmarshal([]byte(val), c)
            if err != nil {
                return
            }
        } else if val, ok := value.(json.RawMessage); ok {
            err = json.Unmarshal([]byte(val), c)
            if err != nil {
                return
            }
        } else if val, ok := value.([]byte); ok {
            err = json.Unmarshal([]byte(val), c)
            if err != nil {
                return
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    case "postgres":
        if val, ok := value.(string); ok {
            val = strings.Trim(val, "{}")
            if val == "" {
                *c = make(ConfirmTypes, 0)
                return
            }
            for _, a := range strings.Split(val, ",") {
                var i int
                i, err = strconv.Atoi(a)
                if err != nil {
                    return
                }
                *c = append(*c, ConfirmType(i))
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    }
    return
}

func (c ConfirmTypes) Value() (value driver.Value, err error) {
    data := ""
    for _, a := range c {
        data = fmt.Sprintf("%s%d,", data, a)
    }
    data = strings.TrimRight(data, ",")
    switch config.DBservice {
    case "mysql", "sqlite":
        value = datatypes.JSON([]byte(fmt.Sprintf("[%s]", data)))
        err = nil
    case "postgres":
        value = fmt.Sprintf("{%s}", data)
        err = nil
    }
    return
}

func (ConfirmTypes) GormDataType() string {
    switch config.DBservice {
    case "mysql", "sqlite":
	    return "json"
    case "postgres":
        return "smallint[]"
    }
    return ""
}

func (ConfirmTypes) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "smallint[]"
	}
	return ""
}

func (js ConfirmTypes) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
	switch db.Dialector.Name() {
    case "sqlite":
        if len(js) == 0 {
            expr = gorm.Expr("NULL")
            return
        }
        data, _ := js.Value()
        expr = gorm.Expr("?", string(data.(datatypes.JSON)))
    case "mysql":
        if len(js) == 0 {
            expr = gorm.Expr("NULL")
            return
        }
        data, _ := js.Value()
        if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
            expr = gorm.Expr("CAST(? AS JSON)", string(data.(datatypes.JSON)))
            return
        }
        expr = gorm.Expr("?", string(data.(datatypes.JSON)))
    case "postgres":
        data, _ := js.Value()
        expr = gorm.Expr("?", data)
	}
    return
}
