//go:generate go run ../../gencode.go struct.tmpl config.yml .

package lab

import (
    "fmt"
    "os"
    "path"
    "bytes"

    "gopkg.in/yaml.v3"
    "golang.org/x/exp/slices"

    "github.com/NCKU-NASA/nasa-judge-lib/schema/user"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/deadline"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/command"
    checkpoint "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/check_point"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/content"

    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
    "github.com/NCKU-NASA/nasa-judge-lib/utils/instance"
    "github.com/NCKU-NASA/nasa-judge-lib/utils/host"
    workertype "github.com/NCKU-NASA/nasa-judge-lib/enum/worker_type"
    contenttype "github.com/NCKU-NASA/nasa-judge-lib/enum/content_type"
)

const (
    LabDir = "./labs"
)

type Lab struct {
    ID uint `gorm:"primaryKey" yaml:"-" json:"-"`
    LabId string `gorm:"unique" yaml:"-" json:"labId"`
    Promissions []user.Group `gorm:"many2many:lab_promissions" yaml:"promissions" json:"promissions"`
    Deadlines deadline.Deadlines `yaml:"deadlines" json:"deadlines"`
    Network ipv4net `yaml:"network" json:"network"`
    Description string `yaml:"description" json:"description"`
    Init command.Commands `yaml:"init" json:"init"`
    Clear command.Commands `yaml:"clear" json:"clear"`
    CheckPoints checkpoint.CheckPoints `yaml:"checkpoints" json:"checkpoints"`
    Contents content.Contents `yaml:"contents" json:"contents"`
}

func init() {
    database.GetDB().AutoMigrate(&Lab{})
}

func Commit(labId string) error {
    if labId == "all" {
        entries, err := os.ReadDir(LabDir)
        if err != nil {
            return err
        }
        for _, e := range entries {
            if e.IsDir() && len(e.Name()) > 0 && e.Name()[0] != '.' && e.Name() != "all" {
                err = Commit(e.Name())
                if err != nil {
                    return err
                }
            }
        }
        return nil
    }

    var lab Lab
    labyaml, err := os.ReadFile(path.Join(LabDir, labId, "config.yml"))
    if err != nil {
        return nil
    }
    err = yaml.Unmarshal(labyaml, &lab)
    if err != nil {
        return err
    }
    lab.LabId = labId
    _, err = GetLab(labId)
    if err != nil {
        err = nil
        result = database.GetDB().Model(&Lab{}).Preload("Promissions").Create(&lab)
        return result.Error
    } else {
        result := database.GetDB().Model(&Lab{}).Preload("Promissions").Where("lab_id = ?", lab.LabId).Updates(&lab)
        return result.Error
    }
}

func GetLab(labId string) (lab Lab, err error) {
    result := database.GetDB().Model(&Lab{}).Preload("Promissions").Where("lab_id = ?", labId).First(&lab)
    err = result.Error
    return
}

func GetLabs() (labs []Lab, err error) {
    result := database.GetDB().Model(&Lab{}).Preload("Promissions").Find(&labs)
    err = result.Error
    return
}

func (lab Lab) ContainPromission(group string) bool {
    return group == "all" || slices.ContainsFunc(lab.Promissions, func(g user.Group) bool {
        return g.Groupname == group
    })
}

func (lab Lab) Judge(userdata user.User, env map[string]string, contents content.Contents) (result checkpoint.CheckPoints, stdout, stderr bytes.Buffer, err error) {
    dir := path.Join(LabDir, lab.LabId)
    connectinfo := make(map[workertype.WorkerType]map[string]instance.Instance)
    connectinfo[workertype.Host] = make(map[string]instance.Instance)
    connectinfo[workertype.Host][""], err = host.NewHost(dir, contents)
    if err != nil {
        return
    }
    connectinfo[workertype.Host][""].Up()
    
    defer func() {
        for _, nowclear := range lab.Clear {
            nowclear.Run(dir, contents, env, connectinfo)
        }

        for _, instancemap := range connectinfo {
            for _, nowinstance := range instancemap {
                nowinstance.Down()
                nowinstance.Close()
            }
        }
    }()

    for _, nowcontent := range contents {
        if nowcontent.Type == contenttype.Input {
            env[nowcontent.Name] = nowcontent.Data
        }
    }
    env["username"] = userdata.Username
    env["userip"] = lab.Network.Nth(uint32(userdata.ID)).String()
    resulttmp := make(map[string]map[int]checkpoint.CheckPoint)

    for _, nowinit := range lab.Init {
        _, _, runstatus := nowinit.Run(dir, contents, env, connectinfo)
        if !runstatus {
            err = fmt.Errorf("Init fail")
            return
        }
    }
    
    for key, checkpointlist := range lab.CheckPoints {
        for idx, nowcheckpoint := range checkpointlist {
            nowstdout, nowstderr := nowcheckpoint.Run(key, idx, dir, contents, env, connectinfo, lab.CheckPoints, resulttmp)
            stdout.Write(nowstdout.Bytes())
            stderr.Write(nowstderr.Bytes())
        }
    }
    
    result = make(checkpoint.CheckPoints)
    for key, checkpointlist := range lab.CheckPoints {
        result[key] = make([]checkpoint.CheckPoint, len(checkpointlist))
        for idx, _ := range checkpointlist {
            result[key][idx] = resulttmp[key][idx]
        }
    }
    return
}


