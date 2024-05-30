package host

import (
    "regexp"

    "github.com/google/uuid"
)

var idmap map[string]*Host

func init() {
    idmap = make(map[string]*Host)
}

func genid() (id string) {
    reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
    id = reg.ReplaceAllString(uuid.NewString(), "")
    for _, exist := idmap[id]; exist; _, exist = idmap[id] {
        id = reg.ReplaceAllString(uuid.NewString(), "")
    }
    return
}

