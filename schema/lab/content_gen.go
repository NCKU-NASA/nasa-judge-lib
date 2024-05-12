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

    
    contenttype "github.com/NCKU-NASA/nasa-judge-lib/enum/content_type"
    
)

type content struct {
    Type contenttype.ContentType
    Name string
}

type contents []content




func (c *contents) Scan(value interface{}) (err error) {
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

func (c contents) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}

func (contents) GormDataType() string {
    return "json"
}

func (contents) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js contents) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
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
