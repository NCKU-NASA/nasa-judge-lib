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

func (c ipv4net) MarshalYAML() (any, error) {
    value := ""
    if c.IPv4Net != nil {
        value = c.String()
    }
    return value, nil
}

func (c *ipv4net) UnmarshalYAML(b *yaml.Node) error {
    var tmp string
    err := b.Decode(&tmp)
    if err != nil {
        return err
    }
    if tmp == "" {
        *c = ipv4net{IPv4Net: nil}
        return nil
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
    } else if val, ok := value.([]uint8); ok {
        var tmpnet *netaddr.IPv4Net
        tmpnet, err = netaddr.ParseIPv4Net(string(val))
        *c = ipv4net{IPv4Net: tmpnet}
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c ipv4net) Value() (value driver.Value, err error) {
    if c.IPv4Net == nil {
        value = ""
        return
    }
    value = c.String()
    return
}
