package frontend_variable_type

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
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
    
    
    "golang.org/x/exp/slices"
    
)

type FrontendVariableType byte
type FrontendVariableTypes []FrontendVariableType

const (
    Download FrontendVariableType = iota
    Input FrontendVariableType
    Upload FrontendVariableType
    VM FrontendVariableType
    Unknown FrontendVariableType = 255
)
var set []string
func init() {
    set = []string{"Download", "Input", "Upload", "VM"}
}

func (c FrontendVariableType) String() string {
    if c == Unknown {
        return ""
    }
    return set[c]
}

func FromString(s string) FrontendVariableType {
    if s == "" {
        return Unknown
    }
    return FrontendVariableType(byte(slices.Index(set, s)))
}

func Random() FrontendVariableType {
    return FrontendVariableType(rand.Intn(Len()))
}

func Len() int {
    return len(set)
}

func (c FrontendVariableType) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.String())
}

func (c FrontendVariableType) MarshalYAML() ([]byte, error) {
    return yaml.Marshal(c.String())
}

func (c *FrontendVariableType) UnmarshalJSON(b []byte) error {
    var num int
    err := json.Unmarshal(b, &num)
    if err == nil {
        *c = FrontendVariableType(num)
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

func (c *FrontendVariableType) UnmarshalYAML(b []byte) error {
    var num int
    err := yaml.Unmarshal(b, &num)
    if err == nil {
        *c = FrontendVariableType(num)
    } else {
        var tmp string
        err = yaml.Unmarshal(b, &tmp)
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

func (c FrontendVariableTypes) MarshalJSON() ([]byte, error) {
    return json.Marshal([]FrontendVariableType(c))
}

func (c FrontendVariableTypes) MarshalYAML() ([]byte, error) {
    return yaml.Marshal([]FrontendVariableType(c))
}

func (c *FrontendVariableTypes) UnmarshalJSON(b []byte) error {
    var tmp []FrontendVariableType
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = FrontendVariableTypes(tmp)
    return err
}

func (c *FrontendVariableTypes) UnmarshalYAML(b []byte) error {
    var tmp []FrontendVariableType
    err := yaml.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = FrontendVariableTypes(tmp)
    return err
}

func (c *FrontendVariableTypes) Scan(value interface{}) (err error) {
    switch database.DBservice {
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
                *c = make(FrontendVariableTypes, 0)
                return
            }
            for _, a := range strings.Split(val, ",") {
                var i int
                i, err = strconv.Atoi(a)
                if err != nil {
                    return
                }
                *c = append(*c, FrontendVariableType(i))
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    }
    return
}

func (c FrontendVariableTypes) Value() (value driver.Value, err error) {
    data := ""
    for _, a := range c {
        data = fmt.Sprintf("%s%d,", data, a)
    }
    data = strings.TrimRight(data, ",")
    switch database.DBservice {
    case "mysql", "sqlite":
        value = datatypes.JSON([]byte(fmt.Sprintf("[%s]", data)))
        err = nil
    case "postgres":
        value = fmt.Sprintf("{%s}", data)
        err = nil
    }
    return
}

func (FrontendVariableTypes) GormDataType() string {
    switch database.DBservice {
    case "mysql", "sqlite":
	    return "json"
    case "postgres":
        return "smallint[]"
    }
    return ""
}

func (FrontendVariableTypes) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js FrontendVariableTypes) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
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
