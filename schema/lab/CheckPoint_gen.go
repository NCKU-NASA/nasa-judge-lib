package lab

import (
    "fmt"
    "strings"
    "reflect"
    "context"
    "encoding/json"
    "database/sql/driver"
    "gorm.io/datatypes"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "gorm.io/gorm/schema"

    
)

type CheckPoint struct {
    Message string `yaml:"message" json:"message"`
    Weight float32 `yaml:"weight" json:"weight"`
    Commands Commands `yaml:"commands" json:"commands,omitempty"`
    Check map[string]int `yaml:"check" json:"check,omitempty"`
    Correct bool `yaml:"-" json:"correct"`
}

type CheckPoints map[string][]CheckPoint




func (c *CheckPoints) Scan(value interface{}) (err error) {
    if val, ok := value.(datatypes.JSON); ok {
        err = json.Unmarshal([]byte(val), c)
    } else if val, ok := value.(json.RawMessage); ok {
        err = json.Unmarshal([]byte(val), c)
    } else if val, ok := value.([]byte); ok {
        err = json.Unmarshal([]byte(val), c)
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    if *c == nil {
        *c = CheckPoints{}
    }
    return
}

func (c CheckPoints) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}

func (CheckPoints) GormDataType() string {
    return "json"
}

func (CheckPoints) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

func (js CheckPoints) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
    if js == nil {
        js = CheckPoints{}
    }
    data, _ := js.Value()
    if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
        expr = gorm.Expr("CAST(? AS JSON)", string(data.(datatypes.JSON)))
        return
    }
    expr = gorm.Expr("?", string(data.(datatypes.JSON)))
    return
}
