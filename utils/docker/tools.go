package docker

import (
    "regexp"

    "github.com/google/uuid"
    netaddr "github.com/dspinhirne/netaddr-go"
)

type rangedata struct {
    min uint64
    max uint64
}

type subnetqueue []rangedata

var idmap map[string]*Docker

var judgeurl string

var subnetpool subnetqueue

var subnetpoolnet *netaddr.IPv4Net
var prefix uint8

func init() {
    idmap = make(map[string]*Docker)
}

func Init(judge string, net *netaddr.IPv4Net, netprefix uint8) {
    judgeurl = judge
    subnetpoolnet = net
    prefix = netprefix

    subnetpool = subnetqueue{rangedata{
        min: 0,
        max: uint64(net.SubnetCount(uint(prefix)) - 1),
    }}
}

func push(src []rangedata, num uint64) (dst []rangedata) {
    if len(src) > 0 && src[len(src) - 1].min <= src[len(src) - 1].max && src[len(src) - 1].max + 1 == num {
        src[len(src) - 1].max++
        dst = src
    } else if len(src) > 0 && src[len(src) - 1].min >= src[len(src) - 1].max && src[len(src) - 1].max - 1 == num {
        src[len(src) - 1].max--
        dst = src
    } else {
        dst = append(src, rangedata{
            min: num,
            max: num,
        })
    }
    return
}

func pop(src []rangedata) (dst []rangedata, num uint64) {
    num = src[0].min
    if src[0].min == src[0].max {
        dst = (src)[1:]
    } else if src[0].min < src[0].max {
        src[0].min++
        dst = src
    } else if src[0].min > src[0].max {
        src[0].min--
        dst = src
    }
    return
}

func remove(src []rangedata, num uint64) (dst []rangedata) {
    for i, _ := range src {
        if src[i].min == src[i].max && src[i].min == num {
            dst = append(src[:i], src[i+1:]...)
            return
        } else if src[i].min < src[i].max && src[i].min <= num && num <= src[i].max {
            tmp := src[i]
            dst = src[:i]
            if tmp.min < num {
                dst = append(dst, rangedata{
                    min: tmp.min,
                    max: num - 1,
                })
            }
            if num < tmp.max {
                dst = append(dst, rangedata{
                    min: num + 1,
                    max: tmp.max,
                })
            }
            dst = append(dst, src[i+1:]...)
            return
        } else if src[i].min > src[i].max && src[i].min >= num && num >= src[i].max {
            tmp := src[i]
            dst = src[:i]
            if tmp.min > num {
                dst = append(dst, rangedata{
                    min: tmp.min,
                    max: num + 1,
                })
            }
            if num > tmp.max {
                dst = append(dst, rangedata{
                    min: num - 1,
                    max: tmp.max,
                })
            }
            dst = append(dst, src[i+1:]...)
            return
        }
    }
    dst = src
    return
}

func (c *subnetqueue) Push(net *netaddr.IPv4Net) {
    length := subnetpoolnet.Resize(uint(prefix)).Len()
    index := uint64((net.Network().Addr() - subnetpoolnet.Network().Addr()) / length)
    *c = subnetqueue(push([]rangedata(*c), index))
}

func (c *subnetqueue) Pop() (net *netaddr.IPv4Net) {
    tmpqueue, tmpnum := pop(*c)
    num := uint32(tmpnum)
    *c = subnetqueue(tmpqueue)
    net = subnetpoolnet.NthSubnet(uint(prefix), num)
    return
}

func (c *subnetqueue) Remove(net *netaddr.IPv4Net) {
    length := subnetpoolnet.Resize(uint(prefix)).Len()
    index := uint64((net.Network().Addr() - subnetpoolnet.Network().Addr()) / length)
    *c = subnetqueue(remove([]rangedata(*c), index))
}


func genid() (id string) {
    reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
    id = reg.ReplaceAllString(uuid.NewString(), "")
    for _, exist := idmap[id]; exist; _, exist = idmap[id] {
        id = reg.ReplaceAllString(uuid.NewString(), "")
    }
    return
}

func gensubnet() *netaddr.IPv4Net {
    return subnetpool.Pop()
}

