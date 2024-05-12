//go:generate go run ../../gencode.go struct.tmpl config.yml .

package lab

import (
    "os"
    "path"

    "gopkg.in/yaml.v3"
    "golang.org/x/exp/slices"

    "github.com/NCKU-NASA/nasa-judge-lib/schema/user"
    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
)

const (
    LabDir = "./labs"
)

type Lab struct {
    ID uint `gorm:"primaryKey" yaml:"_" json:"-"`
    LabId string `gorm:"unique" yaml:"_" json:"labId"`
    Promissions []user.Group `gorm:"many2many:lab_promissions" yaml:"promissions" json:"promissions"`
    Deadlines deadlines `yaml:"deadlines" json:"deadlines"`
    Timeout duration `yaml:"timeout" json:"timeout"`
    Network ipv4net `yaml:"network" json:"network"`
    Description string `yaml:"description" json:"description"`
    Init commands `yaml:"init" json:"init"`
    Clear commands `yaml:"clear" json:"clear"`
    CheckPoints checkpoints `yaml:"checkpoints" json:"checkpoints"`
    FrontendVariable frontendvariables `yaml:"frontendvariable" json:"frontendvariable"`
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
    result := database.GetDB().Model(&Lab{}).Where("lab_id = ?").Updates(&lab)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected > 0 {
        return nil
    }
    result = database.GetDB().Model(&Lab{}).Create(&lab)
    return result.Error
}

func GetLab(labId string) (lab Lab, err error) {
    result := database.GetDB().Model(&Lab{}).Where("lab_id = ?", labId).First(&lab)
    err = result.Error
    return
}

func GetLabs() (labs []Lab, err error) {
    result := database.GetDB().Model(&Lab{}).Find(&labs)
    err = result.Error
    return
}

func (lab Lab) ContainPromission(group string) bool {
    return slices.ContainsFunc(lab.Promissions, func(g user.Group) bool {
        return g.Groupname == group
    })
}
