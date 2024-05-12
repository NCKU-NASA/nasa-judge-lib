package content_type

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
    
)

type ContentType byte
type ContentTypes []ContentType

const (
    Download ContentType = iota
    Input
    Upload
    VM
    Unknown ContentType = 255
)
var set []string
func init() {
    set = []string{"Download", "Input", "Upload", "VM"}
}

func (c ContentType) String() string {
    if c == Unknown {
        return ""
    }
    return set[c]
}

func FromString(s string) ContentType {
    if s == "" {
        return Unknown
    }
    return ContentType(byte(slices.Index(set, s)))
}

func Random() ContentType {
    return ContentType(rand.Intn(Len()))
}

func Len() int {
    return len(set)
}

func (c ContentType) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.String())
}

func (c ContentType) MarshalYAML() (any, error) {
    return c.String(), nil
}

func (c *ContentType) UnmarshalJSON(b []byte) error {
    var num int
    err := json.Unmarshal(b, &num)
    if err == nil {
        *c = ContentType(num)
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

func (c *ContentType) UnmarshalYAML(b *yaml.Node) error {
    var num int
    err := b.Decode(&num)
    if err == nil {
        *c = ContentType(num)
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

func (c ContentTypes) MarshalJSON() ([]byte, error) {
    return json.Marshal([]ContentType(c))
}

func (c ContentTypes) MarshalYAML() (any, error) {
    return []ContentType(c), nil
}

func (c *ContentTypes) UnmarshalJSON(b []byte) error {
    var tmp []ContentType
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = ContentTypes(tmp)
    return err
}

func (c *ContentTypes) UnmarshalYAML(b *yaml.Node) error {
    var tmp []ContentType
    err := b.Decode(&tmp)
    if err != nil {
        return err
    }
    *c = ContentTypes(tmp)
    return err
}

func (c *ContentTypes) Scan(value interface{}) (err error) {
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
                *c = make(ContentTypes, 0)
                return
            }
            for _, a := range strings.Split(val, ",") {
                var i int
                i, err = strconv.Atoi(a)
                if err != nil {
                    return
                }
                *c = append(*c, ContentType(i))
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    }
    return
}

func (c ContentTypes) Value() (value driver.Value, err error) {
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

func (ContentTypes) GormDataType() string {
    switch config.DBservice {
    case "mysql", "sqlite":
	    return "json"
    case "postgres":
        return "smallint[]"
    }
    return ""
}

func (ContentTypes) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js ContentTypes) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
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
