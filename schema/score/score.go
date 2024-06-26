package score

import (
    "time"
    "encoding/json"

    "golang.org/x/exp/slices"

    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/user"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/content"
    checkpoint "github.com/NCKU-NASA/nasa-judge-lib/schema/lab/check_point"
)

type Score struct {
    ID uint `gorm:"primaryKey" json:"-"`
    UserID uint `json:"-"`
    User *user.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
    LabID uint `json:"-"`
    Lab *lab.Lab `gorm:"foreignKey:LabID" json:"lab,omitempty"`
    Score float32 `gorm:"-" json:"score"`
    Result checkpoint.CheckPoints `json:"result,omitempty"`
    Data content.Contents `json:"data,omitempty"`
    Stdout string `json:"stdout,omitempty"`
    Stderr string `json:"stderr,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
}

type Scores struct {
    Scores []Score
    KeyField string
    ShowFields []string
    UseDeadline bool
}

type ScoreFilter struct {
    LabId string `json:"labId"`
    User user.User `json:"user"`
    Score *float32 `json:"score"`
    UseDeadline bool `json:"usedeadline"`
    ShowFields []string `json:"showfields"`
    KeyField string `json:"keyfield"`
    Max bool `json:"max"`
    Groups []groupfilter `json:"groups"`
}

type groupfilter struct {
    Name string `json:"name"`
    Show bool `json:show`
}

func init() {
    database.GetDB().AutoMigrate(&Score{})
}

func (c Score) ToMap() (result map[string]any) {
    result = make(map[string]any)
    result["id"] = c.ID
    result["user"] = c.User
    result["lab"] = c.Lab
    result["score"] = c.Score
    result["result"] = c.Result
    result["data"] = c.Data
    result["stdout"] = c.Stdout
    result["stderr"] = c.Stderr
    result["createdAt"] = c.CreatedAt
    return
}

func MapToScore(c map[string]any) (result Score) {
    result.ID        = c["id"].(uint)
    result.User      = c["user"].(*user.User)
    result.Lab       = c["lab"].(*lab.Lab)
    result.Score     = c["score"].(float32)
    result.Result    = c["result"].(checkpoint.CheckPoints)
    result.Data      = c["data"].(content.Contents)
    result.Stdout    = c["stdout"].(string)
    result.Stderr    = c["stderr"].(string)
    result.CreatedAt = c["createdAt"].(time.Time)
    return
}

func (c Scores) Contain(score Score) bool {
    return slices.ContainsFunc(c.Scores, func(s Score) bool {
        return s.ID == score.ID
    })
}

func (c Scores) MarshalJSON() ([]byte, error) {
    result := make(map[string]map[string][]Score)
    if c.KeyField == "" {
        c.KeyField = "username"
    }
    for _, score := range c.Scores {
        if result[score.Lab.LabId] == nil {
            result[score.Lab.LabId] = make(map[string][]Score)
        }
        score.CalcScore(c.UseDeadline)
        usermap := score.User.ToMap()
        labdata := *(score.Lab)
        if usermap[c.KeyField].(string) != "" {
            orgscore := score.ToMap()
            score.User = nil
            score.Lab = nil
            score.Result = nil
            score.Data = nil
            score.Stdout = ""
            score.Stderr = ""
            scoremap := score.ToMap()
            for _, showfield := range c.ShowFields {
                scoremap[showfield] = orgscore[showfield]
            }
            result[labdata.LabId][usermap[c.KeyField].(string)] = append(result[labdata.LabId][usermap[c.KeyField].(string)], MapToScore(scoremap))
        }
    }
    return json.Marshal(result)
}

func (c *Scores) UnmarshalJSON(b []byte) error {
    var scoresmap map[string]map[string][]Score
    err := json.Unmarshal(b, &scoresmap)
    if err != nil {
        return err
    }
    usermap := make(map[string]user.User)
    for labId, usermapscore := range scoresmap {
        nowlab, err := lab.GetLab(labId)
        if err != nil {
            return err
        }
        for username, scores := range usermapscore {
            if _, exist := usermap[username]; !exist {
                usermap[username], err = user.GetUser("username = ? or student_id = ? or email = ?", username)
                if err != nil {
                    return err
                }
            }
            for _, score := range scores {
                score.UserID = usermap[username].ID
                score.User = new(user.User)
                *(score.User) = usermap[username]
                score.LabID = nowlab.ID
                score.Lab = new(lab.Lab)
                *(score.Lab) = nowlab
                c.Scores = append(c.Scores, score)
            }
        }
    }
    return nil
}

func (score *Score) CalcScore(usedeadline bool) float32 {
    var basescore float32
    for _, allcheckpoint := range score.Result {
        for _, checkpoint := range allcheckpoint {
            if checkpoint.Correct {
                basescore += checkpoint.Weight
            }
        }
    }

    if usedeadline {
        score.Score = 0
        for _, deadline := range score.Lab.Deadlines {
            if score.CreatedAt.Before(deadline.Time) {
                score.Score = basescore * deadline.Score
                break
            }
        }
    } else {
        score.Score = basescore
    }
    return score.Score
}

func (score *Score) Create() error {
    result := database.GetDB().Model(&Score{}).Preload("User").Preload("Lab").Create(score)
    return result.Error
}

func GetScore(id uint, usedeadline bool) (score Score, err error) {
    result := database.GetDB().Model(&Score{}).Preload("User").Preload("Lab").Where("id = ?", id).First(&score)
    err = result.Error
    if err != nil {
        return
    }
    score.CalcScore(usedeadline)
    return
}

func (c ScoreFilter) GetScores(org Scores) (scores Scores, err error) {
    var userdata user.User
    c.User = user.User{
        Username: c.User.Username,
        StudentId: c.User.StudentId,
        Email: c.User.Email,
    }
    c.User.Fix()
    if c.User.Username != "" || c.User.StudentId != "" || c.User.Email != "" {
        userdata, err = user.GetUser(c.User)
        if err != nil {
            err = nil
            return
        }
    }
    var labdata lab.Lab
    if c.LabId != "" {
        labdata, err = lab.GetLab(c.LabId)
        if err != nil {
            err = nil
            return
        }
    }
    if org.Scores == nil {
        nowscore := Score{
            UserID: userdata.ID,
            User: &userdata,
            LabID: labdata.ID,
            Lab: &labdata,
        }

        req := database.GetDB().Model(&Score{}).Preload("User").Preload("Lab").Where(nowscore)
        result := req.Find(&(scores.Scores))
        if result.Error != nil {
            err = result.Error
            return
        }
    } else {
        if userdata.ID != 0 || labdata.ID != 0 {
            for _, nowscore := range org.Scores {
                if (userdata.ID == 0 || nowscore.User.ID == userdata.ID) && (labdata.ID == 0 || nowscore.Lab.ID == labdata.ID) {
                    scores.Scores = append(scores.Scores, nowscore)
                }
            }
        } else {
            scores.Scores = org.Scores
        }
    }
    scores.KeyField = c.KeyField
    scores.ShowFields = c.ShowFields
    if len(c.Groups) > 0 {
        orgscores := scores.Scores
        for _, nowgroupfilter := range c.Groups {
            if nowgroupfilter.Show {
                for _, score := range orgscores {
                    if score.User.ContainGroup(nowgroupfilter.Name) && !scores.Contain(score) {
                        scores.Scores = append(scores.Scores, score)
                    }
                }
            } else {
                for i := 0; i < len(scores.Scores); i++ {
                    if scores.Scores[i].User.ContainGroup(nowgroupfilter.Name) {
                        scores.Scores = append(scores.Scores[:i], scores.Scores[i+1:]...)
                        i--
                    }
                }
            }
        }
    }
    var tmpscores []Score
    for idx, _ := range scores.Scores {
        scores.Scores[idx].CalcScore(c.UseDeadline)
        if c.Score != nil && scores.Scores[idx].Score == *(c.Score) {
            tmpscores = append(tmpscores, scores.Scores[idx])
        }
    }
    if c.Score != nil {
        scores.Scores = tmpscores
    }
    if c.Max {
        maxmap := make(map[uint]map[uint]Score)
        for _, score := range scores.Scores {
            if maxmap[score.Lab.ID] == nil {
                maxmap[score.Lab.ID] = make(map[uint]Score)
            }
            if score.Score >= maxmap[score.Lab.ID][score.User.ID].Score {
                maxmap[score.Lab.ID][score.User.ID] = score
            }
        }
        scores.Scores = make([]Score, 0)
        for _, usermapscore := range maxmap {
            for _, maxscore := range usermapscore {
                scores.Scores = append(scores.Scores, maxscore)
            }
        }
    }
    return
}

