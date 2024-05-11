package lab

import (
    "fmt"
    "time"
    "reflect"
    "database/sql/driver"

    "gopkg.in/yaml.v3"
)

type duration time.Duration

func (c duration) MarshalYAML() ([]byte, error) {
    return yaml.Marshal(float64(c) / float64(time.Second))
}

func (c *duration) UnmarshalYAML(b []byte) error {
    var tmp float64
    err := yaml.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = duration(tmp * float64(time.Second))
    return nil
}

func (c *duration) Scan(value interface{}) (err error) {
    if val, ok := value.(int64); ok {
        *c = duration(val)
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c duration) Value() (value driver.Value, err error) {
    value = int64(c)
    return
}
