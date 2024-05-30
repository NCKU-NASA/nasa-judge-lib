package command

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
    
    "path"
    
    "time"
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/instance"
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/docker"
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/host"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/worker"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/content"
    
    workertype "github.com/NCKU-NASA/nasa-judge-lib/enum/worker_type"
    
)

type Command struct {
    Exec []string `yaml:"exec" json:"exec"`
    Timeout duration `yaml:"timeout" json:"timeout"`
    Worker worker.Worker `yaml:"worker" json:"worker"`
}

type Commands []Command


func (c Command) Run(dir string, contents content.Contents, env map[string]string, connectinfo map[workertype.WorkerType]map[string]instance.Instance) (stdout, stderr bytes.Buffer, result bool) {
    var err error
    if _, exist := connectinfo[c.Worker.WorkerType]; !exist {
        connectinfo[c.Worker.WorkerType] = make(map[string]instance.Instance)
    }
    if _, exist := connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool]; !exist {
        switch c.Worker.WorkerType {
        case workertype.Host:
            connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool], err = host.NewHost(dir, contents)
            if err != nil {
                result = false
                return
            }
            err = connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool].Up()
            if err != nil {
                result = false
                return
            }
        case workertype.Docker:
            if _, exist2 := connectinfo[workertype.Host][""]; !exist2 {
                connectinfo[workertype.Host][""], err = host.NewHost(dir, contents)
                if err != nil {
                    result = false
                    return
                }
                err = connectinfo[workertype.Host][""].Up()
                if err != nil {
                    result = false
                    return
                }
            }
            if len(connectinfo[c.Worker.WorkerType]) <= 0 {
                connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool], err = docker.NewDocker(dir, path.Join("/tmp", connectinfo[workertype.Host][""].GetID(), "workdir"))
                if err != nil {
                    result = false
                    return
                }
                err = connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool].Up()
                if err != nil {
                    result = false
                    return
                }
            } else {
                for _, v := range connectinfo[c.Worker.WorkerType] {
                    connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool] = v
                    break
                }
            }
        case workertype.SSH:
            result = false
            return
        default: 
            result = false
            return
        }
        err = connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool].InitPool(c.Worker.WorkerPool)
        if err != nil {
            result = false
            return
        }
    }
    stdout, stderr, err = connectinfo[c.Worker.WorkerType][c.Worker.WorkerPool].Exec(c.Worker.WorkerPool, env, time.Duration(c.Timeout), c.Exec...)
    result = err == nil
    return
}

func (c *Commands) Scan(value interface{}) (err error) {
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
        *c = Commands{}
    }
    return
}

func (c Commands) Value() (value driver.Value, err error) {
    var tmp []byte
    tmp, err = json.Marshal(c)
    value = datatypes.JSON(tmp)
    return
}

func (Commands) GormDataType() string {
    return "json"
}

func (Commands) GormDBDataType(db *gorm.DB, field *schema.Field) string {
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

func (js Commands) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
    if js == nil {
        js = Commands{}
    }
    data, _ := js.Value()
    if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
        expr = gorm.Expr("CAST(? AS JSON)", string(data.(datatypes.JSON)))
        return
    }
    expr = gorm.Expr("?", string(data.(datatypes.JSON)))
    return
}
