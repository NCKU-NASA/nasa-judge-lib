package lab

import (
    "reflect"
    "encoding/json"
    "database/sql/driver"
    "gorm.io/datatypes"

    
    "github.com/NCKU-NASA/nasa-judge-lib/enum/frontend_variable_type"
    
)

type frontendvariable struct {
    Type frontendvariabletype.FrontendVariableType
    Name string
}

type frontendvariables []frontendvariable




func (c *frontendvariables) Scan(value interface{}) (err error) {
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

func (c frontendvariables) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}
