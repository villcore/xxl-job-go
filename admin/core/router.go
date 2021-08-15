package core

import (
	"errors"
	"villcore.com/common/api"
)

type RouterType string

const (
	ROUTER_FIRST                 RouterType = "FIRST"
	ROUTER_LAST                  RouterType = "LAST"
	ROUTER_ROUND                 RouterType = "ROUND"
	ROUTER_RANDOM                RouterType = "RANDOM"
	ROUTER_CONSISTENT_HASH       RouterType = "CONSISTENT_HASH"
	ROUTER_LEAST_FREQUENTLY_USED RouterType = "LEAST_FREQUENTLY_USED"
	ROUTER_LEAST_RECENTLY_USED   RouterType = "LEAST_RECENTLY_USED"
	ROUTER_FAILOVER              RouterType = "FAILOVER"
	ROUTER_BUSYOVER              RouterType = "BUSYOVER"
	ROUTER_SHARDING_BROADCAST    RouterType = "SHARDING_BROADCAST"
)

func Route(routerType RouterType, triggerParam *api.TriggerParam, addressList []string) (string, error) {
	switch routerType {
	case ROUTER_FIRST:
		return addressList[0], nil
	case ROUTER_LAST:
		return addressList[len(addressList)-1], nil

		// TODO
	case ROUTER_ROUND:
		return "", nil
	case ROUTER_RANDOM:
		return "", nil
	case ROUTER_CONSISTENT_HASH:
		return "", nil
	case ROUTER_LEAST_FREQUENTLY_USED:
		return "", nil
	case ROUTER_LEAST_RECENTLY_USED:
		return "", nil
	case ROUTER_FAILOVER:
		return "", nil
	case ROUTER_BUSYOVER:
		return "", nil
	case ROUTER_SHARDING_BROADCAST:
		return "", nil
	default:
		return "", errors.New("Invalid router type " + string(routerType))
	}
}
