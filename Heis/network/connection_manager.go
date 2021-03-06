package network

import (
	"fmt"
	"net"
	"sync"

	"../types"
)

type connection struct {
	id        int
	sendMsg   chan *types.NetworkMessage
	terminate chan bool
}

type connManager struct {
	masterIP string
	currId   int
	conns    map[*net.TCPConn]*connection

	wakeRecieve chan *types.NetworkMessage
	newConn     chan *net.TCPConn
	connEnd     chan *net.TCPConn

	hubRecieve chan *types.NetworkMessage
	hubSend    chan *types.NetworkMessage

	wg *sync.WaitGroup
}

func newConnManager(hbRec, hbSend chan *types.NetworkMessage) *connManager {
	return &connManager{
		masterIP: "",
		currId:   1,
		conns:    make(map[*net.TCPConn]*connection),

		// buffer for messages recieved
		wakeRecieve: make(chan *types.NetworkMessage, BUFFER_MSG_RECIEVED),
		newConn:     make(chan *net.TCPConn),
		connEnd:     make(chan *net.TCPConn),

		hubRecieve: hbRec,
		hubSend:    hbSend,

		wg: new(sync.WaitGroup),
	}
}

func (cm *connManager) run() {
	fmt.Println("\x1b[34;1m::: Start Connection Manager :::\x1b[0m")

	go startTCPListener(cm.newConn)
	for {
		// prioritized channels to check
		select {
		case conn := <-cm.connEnd:
			cm.removeConnection(conn)
			continue
		case conn := <-cm.newConn:
			cm.addConnection(conn)
			continue
		default:
		}

		select {
		case conn := <-cm.connEnd:
			cm.removeConnection(conn)
		case conn := <-cm.newConn:
			cm.addConnection(conn)
		case recieveMsg := <-cm.wakeRecieve:
			cm.hubRecieve <- recieveMsg
		case sendMsg := <-cm.hubSend:
			numConns := len(cm.conns)
			if numConns > 0 {
				for _, c := range cm.conns {
					msgHolder := new(types.NetworkMessage)
					types.DeepCopy(msgHolder, sendMsg)
					msgHolder.Id = c.id
					cm.wg.Add(1)
					select {
					case c.sendMsg <- msgHolder:
					default:
						cm.wg.Done()
					}
				}
				cm.wg.Wait()
			}
		}
	}
}

func (cm *connManager) connectToNetwork(masterIP string) error {
	fmt.Printf("\x1b[36;1m::: Connecting To Network, Master Ip=%v :::\x1b[0m\n", masterIP)
	cm.masterIP = masterIP
	conn, err := createTCPConn(cm.masterIP)
	if err != nil {
		return err
	}
	cm.addConnection(conn)
	return nil
}

func (cm *connManager) addConnection(conn *net.TCPConn) {
	fmt.Printf("\x1b[36;1m::: Adding New Connection With Id=%v :::\x1b[0m\n", cm.currId)
	c := &connection{
		id:        cm.currId,
		sendMsg:   make(chan *types.NetworkMessage),
		terminate: make(chan bool, 1),
	}
	cm.conns[conn] = c
	cm.currId++
	go runTCPHandler(conn, cm.wakeRecieve, c.sendMsg, cm.connEnd, c.terminate, cm.wg)
}

func (cm *connManager) removeConnection(conn *net.TCPConn) {
	if removeConn, ok := cm.conns[conn]; ok {
		fmt.Printf("\x1b[31;1m::: Removing Connection With Id=%v :::\x1b[0m\n", cm.currId-1)
		delete(cm.conns, conn)
		for conn, c := range cm.conns {
			if c.id > removeConn.id {
				cm.conns[conn].id--
			}
		}
		cm.currId--
	} else {
		fmt.Println("\t\x1b[31;1mError\x1b[0m |cm.removeConnection|",
			"[Did not find a connection to remove in the connection list]")
	}
}
