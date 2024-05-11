package lab

import (
    "fmt"
    "reflect"
    "database/sql/driver"

    "gopkg.in/yaml.v3"
    netaddr "github.com/dspinhirne/netaddr-go"
)

type ipv4net struct { 
    *netaddr.IPv4Net
}

func (c ipv4net) MarshalYAML() ([]byte, error) {
    return yaml.Marshal(c.String())
}

func (c *ipv4net) UnmarshalYAML(b []byte) error {
    var tmp string
    err := yaml.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    tmpnet, err := netaddr.ParseIPv4Net(tmp)
    *c = ipv4net{IPv4Net: tmpnet}
    return err
}

func (c *ipv4net) Scan(value interface{}) (err error) {
    if val, ok := value.(string); ok {
        var tmpnet *netaddr.IPv4Net
        tmpnet, err = netaddr.ParseIPv4Net(val)
        *c = ipv4net{IPv4Net: tmpnet}
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c ipv4net) Value() (value driver.Value, err error) {
    value = c.String()
    return
}
