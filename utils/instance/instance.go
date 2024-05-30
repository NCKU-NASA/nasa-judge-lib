package instance

import (
    "bytes"
    "time"
)

type Instance interface {
    GetID() string
    Up() error
    Down() error
    InitPool(string) error
    Exec(string, map[string]string, time.Duration, ...string) (bytes.Buffer, bytes.Buffer, error)
    Close()
}
