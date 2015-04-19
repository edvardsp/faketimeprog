package network

import (
	"fmt"

	"../types"
)

type netStatManager struct {
	netstat *types.NetworkMessage

	newMsg chan *types.NetworkMessage
	update chan *types.NetworkMessage
	tick   chan bool
}

func newNetStatManager(newMsgCh, updateCh chan *types.NetworkMessage,
	tickCh chan bool) *netStatManager {
	ns := &netStatManager{
		netstat: types.NewNetworkMessage(),

		newMsg: newMsgCh,
		update: updateCh,
		tick:   tickCh,
	}
	ns.netstat.Id = 0
	return ns
}

func (ns *netStatManager) run() {
	for {
		select {
		case newMsg := <-ns.newMsg:
			ns.parseNewMsg(newMsg)
		case <-ns.tick:
			ns.sendUpdate()
		}
	}
}

func (ns *netStatManager) parseNewMsg(msg *types.NetworkMessage) {
	id := msg.Id
	ns.netstat.Statuses[id] = msg.Statuses[id]
	for order := range msg.Orders {
		if order.Completed {
			ns.deleteOrder(id, &order)
		} else {
			ns.addOrder(id, &order)
		}
	}
}

func (ns *netStatManager) addOrder(id int, order *types.Order) {
	if order.ButtonPress == types.BUTTON_INTERNAL {
		newEtg := order.Floor
		internal := ns.netstat.Statuses[id].InternalOrders
		for i, etg := range internal {
			if newEtg == etg {
				break
			} else if etg == -1 {
				internal[i] = newEtg
				return
			}
		}
	} else {
		if _, ok := ns.netstat.Orders[*order]; !ok {
			ns.netstat.Orders[*order] = struct{}{}
		}
	}
}

func (ns *netStatManager) deleteOrder(id int, order *types.Order) {
	if order.ButtonPress == types.BUTTON_INTERNAL {
		internal := ns.netstat.Statuses[id].InternalOrders
		internal = append(internal[1:], -1)
		elevstat := ns.netstat.Statuses[id]
		elevstat.InternalOrders = internal
		ns.netstat.Statuses[id] = elevstat
	} else {
		order.Completed = false
		delete(ns.netstat.Orders, *order)
	}
}

func (ns *netStatManager) sendUpdate() {
	fmt.Println("netstat ::", ns.netstat)
	nm := types.NewNetworkMessage()
	types.Clone(nm, ns.netstat)
	ns.update <- nm
	for id := range ns.netstat.Statuses {
		if id != 0 {
			delete(ns.netstat.Statuses, id)
		}
	}
}
