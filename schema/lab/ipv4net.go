package lab

import (
    netaddr "github.com/dspinhirne/netaddr-go"
)

type ipv4net *netaddr.IPv4Net

func (c ipv4net) MarshalYAML() ([]byte, error) {
    return yaml.Marshal(c.String())
}

func (c *ipv4net) UnmarshalYAML(b []byte) error {
    var tmp string
    err := yaml.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c, err = ParseIPv4Net(tmp)
    return err
}

func (c *ipv4net) Scan(value interface{}) (err error) {
    if val, ok := value.(string); ok {
        *c, err = ParseIPv4Net(val)
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c ipv4net) Value() (value driver.Value, err error) {
    value = c.String()
    return
}
