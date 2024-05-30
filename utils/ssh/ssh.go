package ssh

import (
    "os"
    "path"
    "golang.org/x/crypto/ssh"
)

type SSH struct {
    IP *netaddr.IPv4
    PoolName string
    Dir string
    session *ssh.Session
}

var ipmap map[string]*SSH

var lock *sync.RWMutex

var worker map[string][]string

func init() {
    lock = new(sync.RWMutex)
}

func Init(confworker map[string][]string) {
    worker = confworker
}

func (c *SSH) Up() error {
    lock.RLock()
    if data, exist := ipmap[c.IP.String()]; !exist || data != c || c.session != nil {
        lock.RUnlock()
        return fmt.Errorf("Use after free")
    }
    lock.RUnlock()
    home, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    privateBytes, err := os.ReadFile(path.join(home, ".ssh/id_rsa"))
    if err != nil {
        return err
    }
    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil {
        return err
    }
    config := &ssh.ClientConfig{
        User: username,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(private),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", c.IP.String()), config)
    if err != nil {
        return err
    }
    defer conn.Close()
    c.session, err = conn.NewSession()
    return err
}

func (c *SSH) Down() error {
    lock.RLock()
    if data, exist := ipmap[c.IP.String()]; !exist || data != c || c.session == nil {
        lock.RUnlock()
        return fmt.Errorf("Use after free")
    }
    lock.RUnlock()
    c.session.Close()
    c.session = nil
}

func (c *SSH) Exec() (stdout, stderr bytes.Buffer, err error) {
    lock.RLock()
    if data, exist := ipmap[c.IP.String()]; !exist || data != c || c.session == nil {
        lock.RUnlock()
        err = fmt.Errorf("Use after free")
        return
    }
    lock.RUnlock()
    session.Stdout = &stdout
    session.Stderr = &stderr
    for k, v := range env {
        c.Setenv(k, v)
    }
    err := session.Run()
}

func NewSSH(name, dir string) (*SSH, error) {
    lock.Lock()
    defer lock.Unlock()
    if len(worker[name]) <= 0 {
        return nil, fmt.Errorf("Pool Busy")
    }
    ins := SSH{
        IP: netaddr.ParseIPv4(worker[name][0]),
        PoolName: name,
        Dir: dir,
    }
    worker[name] = worker[name][1:]
    return &ins, nil
}

func (c *SSH) Close() {
    lock.Lock()
    defer lock.Unlock()
    if data, exist := ipmap[c.IP.String()]; exist && data == c {
        delete(ipmap, c.IP.String())
        worker[c.PoolName] = append(worker[c.PoolName], c.IP.String())
    }
}
