package check_point

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

    
    "bytes"
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/instance"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/command"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/content"
    
    workertype "github.com/NCKU-NASA/nasa-judge-lib/enum/worker_type"
    
)

type CheckPoint struct {
    Message string `yaml:"message" json:"message"`
    Weight float32 `yaml:"weight" json:"weight"`
    Commands command.Commands `yaml:"commands" json:"commands,omitempty"`
    Dependencies map[string][]int `yaml:"dependencies" json:"dependencies,omitempty"`
    Correct bool `yaml:"-" json:"correct"`
}

type CheckPoints map[string][]CheckPoint


func (c CheckPoint) Run(name string, index int, dir string, contents content.Contents, env map[string]string, connectinfo map[workertype.WorkerType]map[string]instance.Instance, checkpoints CheckPoints, result map[string]map[int]CheckPoint) (stdout, stderr bytes.Buffer) {
    if _, exist := result[name]; !exist {
        result[name] = make(map[int]CheckPoint)
    }
    if _, exist := result[name][index]; exist {
        return
    }
    for k, v := range c.Dependencies {
        if _, exist := result[k]; !exist {
            result[k] = make(map[int]CheckPoint)
        }
        for _, num := range v {
            if _, exist := result[k][num]; !exist {
                nowstdout, nowstderr := checkpoints[k][num].Run(k, num, dir, contents, env, connectinfo, checkpoints, result)
                stdout.Write(nowstdout.Bytes())
                stderr.Write(nowstderr.Bytes())
            }
            if !result[k][num].Correct {
                tmp := c
                tmp.Commands = nil
                tmp.Dependencies = nil
                tmp.Correct = false
                result[name][index] = tmp
                return
            }
        }
    }
    stdout.WriteString(fmt.Sprintf("%s/%d\n", name, index))
    stderr.WriteString(fmt.Sprintf("%s/%d\n", name, index))
    tmp := c
    tmp.Commands = nil
    tmp.Dependencies = nil
    tmp.Correct = true
    for _, command := range c.Commands {
        nowstdout, nowstderr, status := command.Run(dir, contents, env, connectinfo)
        stdout.Write(nowstdout.Bytes())
        stderr.Write(nowstderr.Bytes())
        if !status {
            tmp.Correct = false
            break
        }
    }
    result[name][index] = tmp
    stdout.WriteString("\n")
    stderr.WriteString("\n")
    return
}

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
