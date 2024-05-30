package host

import (
    "fmt"
    "bytes"
    "os"
    "os/exec"
    "path"
    "sync"
    "text/template"
    "encoding/base64"
    "context"
    "time"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/content"
    contenttype "github.com/NCKU-NASA/nasa-judge-lib/enum/content_type"
)

type Host struct {
    ID string
    Dir string
    contents content.Contents
}

var lock *sync.RWMutex
var judgeurl string

func init() {
    lock = new(sync.RWMutex)
}

func Init(judge string) {
    judgeurl = judge
}

func (c *Host) GetID() string {
    return c.ID
}

func (c *Host) Up() error {
    lock.Lock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        lock.Unlock()
        return fmt.Errorf("Use after free")
    }
    lock.Unlock()
    err := os.Mkdir(path.Join("/tmp", c.ID), os.ModePerm)
    if err != nil {
        return err
    }
    err = os.Mkdir(path.Join("/tmp", c.ID, "workdir"), os.ModePerm)
    if err != nil {
        return err
    }
    for _, nowcontent := range c.contents {
        if nowcontent.Type == contenttype.Upload {
            fileData, err := base64.StdEncoding.DecodeString(nowcontent.Data)
            if err != nil {
                return err
            }
            err = os.WriteFile(path.Join("/tmp", c.ID, "workdir", nowcontent.Name), fileData, 0644)
            if err != nil {
                return err
            }
        }
    }
    err = exec.Command("cp", "-r", c.Dir, path.Join("/tmp", c.ID, "lab")).Run()
    return err
}

func (c *Host) Down() error {
    lock.Lock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        lock.Unlock()
        return fmt.Errorf("Use after free")
    }
    lock.Unlock()
    err := exec.Command("rm", "-rf", path.Join("/tmp", c.ID)).Run()
    return err
}

func (c *Host) InitPool(poolname string) error {
    return nil
}

func (c *Host) Exec(name string, env map[string]string, timeout time.Duration, command ...string) (stdout, stderr bytes.Buffer, err error) {
    lock.Lock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        lock.Unlock()
        err = fmt.Errorf("Use after free")
        return
    }
    lock.Unlock()
    env["JUDGEURL"] = judgeurl
    env["LABDIR"] = path.Join("/tmp", c.ID, "lab")
    t := template.New("").Funcs(template.FuncMap{})
    for i, d := range command {
        nowt := template.Must(t.Parse(d))
        var buf bytes.Buffer
        nowt.Execute(&buf, env)
        command[i] = buf.String()
    }
    fmt.Println(command)
    ctx := context.Background()
    if timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(context.Background(), timeout)
        defer cancel()
    }
    cmd := exec.CommandContext(ctx, command[0], command[1:]...)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    cmd.Dir = path.Join("/tmp", c.ID, "workdir")
    for k, v := range env {
        cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
    }
    err = cmd.Run()
    return
}

func NewHost(dir string, contents content.Contents) (*Host, error) {
    lock.Lock()
    defer lock.Unlock()
    ins := Host{
        ID: genid(),
        Dir: dir,
        contents: contents,
    }
    idmap[ins.ID] = &ins
    return &ins, nil
}

func (c *Host) Close() {
    lock.Lock()
    defer lock.Unlock()
    if data, exist := idmap[c.ID]; exist && data == c {
        delete(idmap, c.ID)
    }
}
