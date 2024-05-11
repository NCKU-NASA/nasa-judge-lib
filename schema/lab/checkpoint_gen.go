package lab

import (
    "fmt"
    "reflect"
    "encoding/json"
    "database/sql/driver"
    "gorm.io/datatypes"

    
)

type checkpoint struct {
    Message string `yaml:"message" json:"message"`
    Weight float32 `yaml:"weight" json:"weight"`
    Commands commands `yaml:"commands" json:"commands"`
    Check map[string]int `yaml:"check" json:"check"`
}

type checkpoints []checkpoint




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
