package lab

import (
    "fmt"
    "reflect"
    "encoding/json"
    "database/sql/driver"
    "gorm.io/datatypes"

    
)

type command struct {
    Exec []string `yaml:"exec" json:"exec"`
    Worker worker `yaml:"worker" json:"worker"`
}

type commands []command




func (c *commands) Scan(value interface{}) (err error) {
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

func (c commands) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}
