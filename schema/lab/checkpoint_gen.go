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

type checkpoint struct {
    Message string `yaml:"message" json:"message"`
    Weight float32 `yaml:"weight" json:"weight"`
    Commands commands `yaml:"commands" json:"commands"`
    Check map[string]int `yaml:"check" json:"check"`
}

type checkpoints map[string][]checkpoint




func (c *checkpoints) Scan(value interface{}) (err error) {
    if val, ok := value.(datatypes.JSON); ok {
        err = json.Unmarshal([]byte(val), c)
    } else if val, ok := value.(json.RawMessage); ok {
        err = json.Unmarshal([]byte(val), c)
    } else if val, ok := value.([]byte); ok {
        err = json.Unmarshal([]byte(val), c)
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c checkpoints) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}

func (checkpoints) GormDataType() string {
    return "json"
}

func (checkpoints) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js checkpoints) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
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
    return
}
