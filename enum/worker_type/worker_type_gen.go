package worker_type

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

type WorkerType byte
type WorkerTypes []WorkerType

const (
    Host WorkerType = iota
    Docker WorkerType
    VM WorkerType
    Unknown WorkerType = 255
)
var set []string
func init() {
    set = []string{"Host", "Docker", "VM"}
}

func (c WorkerType) String() string {
    if c == Unknown {
        return ""
    }
    return set[c]
}

func FromString(s string) WorkerType {
    if s == "" {
        return Unknown
    }
    return WorkerType(byte(slices.Index(set, s)))
}

func Random() WorkerType {
    return WorkerType(rand.Intn(Len()))
}

func Len() int {
    return len(set)
}

func (c WorkerType) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.String())
}

func (c WorkerType) MarshalYAML() ([]byte, error) {
    return yaml.Marshal(c.String())
}

func (c *WorkerType) UnmarshalJSON(b []byte) error {
    var num int
    err := json.Unmarshal(b, &num)
    if err == nil {
        *c = WorkerType(num)
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

func (c *WorkerType) UnmarshalYAML(b []byte) error {
    var num int
    err := yaml.Unmarshal(b, &num)
    if err == nil {
        *c = WorkerType(num)
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

func (c WorkerTypes) MarshalJSON() ([]byte, error) {
    return json.Marshal([]WorkerType(c))
}

func (c WorkerTypes) MarshalYAML() ([]byte, error) {
    return yaml.Marshal([]WorkerType(c))
}

func (c *WorkerTypes) UnmarshalJSON(b []byte) error {
    var tmp []WorkerType
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = WorkerTypes(tmp)
    return err
}

func (c *WorkerTypes) UnmarshalYAML(b []byte) error {
    var tmp []WorkerType
    err := yaml.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = WorkerTypes(tmp)
    return err
}

func (c *WorkerTypes) Scan(value interface{}) (err error) {
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
                *c = make(WorkerTypes, 0)
                return
            }
            for _, a := range strings.Split(val, ",") {
                var i int
                i, err = strconv.Atoi(a)
                if err != nil {
                    return
                }
                *c = append(*c, WorkerType(i))
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    }
    return
}

func (c WorkerTypes) Value() (value driver.Value, err error) {
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

func (WorkerTypes) GormDataType() string {
    switch database.DBservice {
    case "mysql", "sqlite":
	    return "json"
    case "postgres":
        return "smallint[]"
    }
    return ""
}

func (WorkerTypes) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js WorkerTypes) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
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
