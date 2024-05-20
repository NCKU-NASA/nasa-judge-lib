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

    
    "time"
    
    "gopkg.in/yaml.v3"
    
)

type Deadline struct {
    Time time.Time `yaml:"time" json:"time"`
    Score float32 `yaml:"score" json:"score"`
}

type Deadlines []Deadline


const (
    format = "2006-01-02 15:04:05"
)

func (c Deadline) MarshalYAML() (any, error) {
    return struct {
        Time string `yaml:"time"`
        Score float32 `yaml:"score"`
    }{
        Time: c.Time.Format(format),
        Score: c.Score,
    }, nil
}

func (c *Deadline) UnmarshalYAML(b *yaml.Node) error {
    var tmp struct {
        Time string `yaml:"time"`
        Score float32 `yaml:"score"`
    }
    err := b.Decode(&tmp)
    if err != nil {
        return err
    }
    if tmp.Time == "" {
        tmp.Time = "9999-12-31 23:59:59"
    }
    tmptime, err := time.ParseInLocation(format, tmp.Time, time.Local)
    if err != nil {
        return err
    }
    *c = Deadline{
        Time: tmptime,
        Score: c.Score,
    }
    return nil
}

func (c *Deadlines) Scan(value interface{}) (err error) {
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
        *c = Deadlines{}
    }
    return
}

func (c Deadlines) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}

func (Deadlines) GormDataType() string {
    return "json"
}

func (Deadlines) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js Deadlines) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
    if js == nil {
        js = Deadlines{}
    }
    data, _ := js.Value()
    if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
        expr = gorm.Expr("CAST(? AS JSON)", string(data.(datatypes.JSON)))
        return
    }
    expr = gorm.Expr("?", string(data.(datatypes.JSON)))
    return
}
