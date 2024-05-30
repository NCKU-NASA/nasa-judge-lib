package docker

import (
    "fmt"
    "bytes"
    "regexp"
    "os"
    "os/exec"
    "path"
    "sync"
    "text/template"
    "context"
    "time"
    netaddr "github.com/dspinhirne/netaddr-go"
)

type Docker struct {
    ID string
    SubNets netaddr.IPv4NetList
    Dir string
    hostworkerdir string
}

var lock *sync.RWMutex

func init() {
    lock = new(sync.RWMutex)
}

func (c *Docker) GetID() string {
    return c.ID
}

func (c *Docker) Up() error {
    lock.RLock()
    defer lock.RUnlock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        return fmt.Errorf("Use after free")
    }
    cmd := exec.Command("docker", "compose", "-p", c.ID, "up", "-d")
    cmd.Dir = c.Dir
    for i, subnet := range c.SubNets {
        cmd.Env = append(cmd.Environ(), fmt.Sprintf("SUBNET%d=%s", i, subnet.String()))
    }
    return cmd.Run()
}

func (c *Docker) Down() error {
    lock.RLock()
    defer lock.RUnlock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        return fmt.Errorf("Use after free")
    }
    cmd := exec.Command("docker", "compose", "-p", c.ID, "down")
    cmd.Dir = c.Dir
    for i, subnet := range c.SubNets {
        cmd.Env = append(cmd.Environ(), fmt.Sprintf("SUBNET%d=%s", i, subnet.String()))
    }
    return cmd.Run()
}

func (c *Docker) InitPool(poolname string) error {
    lock.RLock()
    defer lock.RUnlock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        return fmt.Errorf("Use after free")
    }
    entries, err := os.ReadDir(c.hostworkerdir)
    if err != nil {
        return err
    }
    for _, entry := range entries {
        nowcommand := []string{"docker", "compose", "-p", c.ID, "cp", path.Join(c.hostworkerdir, entry.Name()), fmt.Sprintf("%s:/workdir", poolname)}
        fmt.Println(nowcommand)
        cmd := exec.Command(nowcommand[0], nowcommand[1:]...)
        cmd.Dir = c.Dir
        for i, subnet := range c.SubNets {
            cmd.Env = append(cmd.Env, fmt.Sprintf("SUBNET%d=%s", i, subnet.String()))
        }
        err = cmd.Run()
        if err != nil {
            return err
        }
    }
    return nil
}

func (c *Docker) Exec(name string, env map[string]string, timeout time.Duration, command ...string) (stdout, stderr bytes.Buffer, err error) {
    lock.RLock()
    if data, exist := idmap[c.ID]; !exist || data != c {
        lock.RUnlock()
        err = fmt.Errorf("Use after free")
        return
    }
    lock.RUnlock()
    env["JUDGEURL"] = judgeurl
    env["LABDIR"] = "/lab"
    t := template.New("").Funcs(template.FuncMap{})
    for i, d := range command {
        nowt := template.Must(t.Parse(d))
        var buf bytes.Buffer
        nowt.Execute(&buf, env)
        command[i] = buf.String()
    }
    nowcommand := []string{"docker", "compose", "-p", c.ID, "exec"}
    for k, v := range env {
        nowcommand = append(nowcommand, "-e")
        nowcommand = append(nowcommand, fmt.Sprintf("%s=%s", k, v))
    }
    nowcommand = append(nowcommand, name)
    nowcommand = append(nowcommand, command...)
    fmt.Println(nowcommand)
    ctx := context.Background()
    if timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(context.Background(), timeout)
        defer cancel()
    }
    cmd := exec.CommandContext(ctx, nowcommand[0], nowcommand[1:]...)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    cmd.Dir = c.Dir
    for i, subnet := range c.SubNets {
        cmd.Env = append(cmd.Environ(), fmt.Sprintf("SUBNET%d=%s", i, subnet.String()))
    }
    err = cmd.Run()
    return
}

func NewDocker(dir, hostworkerdir string) (*Docker, error) {
    lock.Lock()
    defer lock.Unlock()
    reg := regexp.MustCompile(`\$\{SUBNET[0-9]+\}`)
    composefile, err := os.ReadFile(path.Join(dir, "docker-compose.yml"))
    if err != nil {
        return nil, err
    }
    subnetlength := len(reg.FindAllString(string(composefile), -1))
    if len(subnetpool) < subnetlength {
        return nil, fmt.Errorf("Pool Busy")
    }
    subnets := make(netaddr.IPv4NetList, subnetlength)
    for i := 0; i < subnetlength; i++ {
        subnets[i] = gensubnet()
    }
    ins := Docker{
        ID: genid(),
        SubNets: subnets,
        Dir: dir,
        hostworkerdir: hostworkerdir,
    }
    idmap[ins.ID] = &ins
    return &ins, nil
}

func (c *Docker) Close() {
    lock.Lock()
    defer lock.Unlock()
    if data, exist := idmap[c.ID]; exist && data == c {
        delete(idmap, c.ID)
        for _, subnet := range c.SubNets {
            subnetpool.Push(subnet)
        }
    }
}
