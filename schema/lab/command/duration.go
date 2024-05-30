package command

import (
    "time"
    "encoding/json"

    "gopkg.in/yaml.v3"
)

type duration time.Duration

func (c duration) MarshalYAML() (any, error) {
    return float64(c) / float64(time.Second), nil
}

func (c *duration) UnmarshalYAML(b *yaml.Node) error {
    var tmp float64
    err := b.Decode(&tmp)
    if err != nil {
        return err
    }
    *c = duration(tmp * float64(time.Second))
    return nil
}

func (c duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(float64(c) / float64(time.Second))
}

func (c *duration) UnmarshalJSON(b []byte) error {
    var tmp float64
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = duration(tmp * float64(time.Second))
    return nil
}
