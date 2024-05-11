package lab

import (
    "reflect"
    "encoding/json"
    "database/sql/driver"
    "gorm.io/datatypes"

    
    "time"
    
    "gopkg.in/yaml.v3"
    
)

type deadline struct {
    Time time.Time `yaml:"time" json:"time"`
    Score float32 `yaml:"score" json:"score"`
}

type deadlines []deadline


const (
    format = "2006-01-02 15:04:05"
)

func (c deadline) MarshalYAML() ([]byte, error) {
    return yaml.Marshal(struct {
        Time string `yaml:"time"`
        Score float32 `yaml:"score"`
    }{
        Time: c.Time.Format(format)
        Score: c.Score
    })
}

func (c *deadline) UnmarshalYAML(b []byte) error {
    var tmp struct {
        Time string `yaml:"time"`
        Score float32 `yaml:"score"`
    }
    err := yaml.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    tmptime, err = time.ParseInLocation(format, tmp.Time, time.Local)
    if err != nil {
        return err
    }
    *c = deadline{
        Time: tmptime,
        Score: c.Score,
    }
    return nil
}

func (c *deadlines) Scan(value interface{}) (err error) {
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

func (c deadlines) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}