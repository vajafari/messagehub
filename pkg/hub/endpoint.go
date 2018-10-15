package hub

// وظیفه ی این کلاس این است که یک سرور را استارت کند.
// و وقتی یک کانکشن جدید ایجاد شد و یا بسته شد به ماژول کنترل کننده خبر دهد

import (
	"log"
	"math/rand"
	"net"
	"strconv"

	"github.com/vajafari/messagehub/pkg/socket"
)

// EndpointConfing Containt configuration for tcp end point
type EndpointConfing struct {
	Host          string
	Port          int
	NetType       string
	SendQueueSize int
	ReadBufSize   int
	WriteBufSize  int
}

// GetHostAddress Apprend host address and port number together and return  full address of site
func (conf *EndpointConfing) GetHostAddress() string {
	return conf.Host + ":" + strconv.Itoa(conf.Port)
}

// Endpoint is tcp endpint that handle input connections
type Endpoint struct {
	config   EndpointConfing  // server configuration
	listener *net.TCPListener // reference to listener
	// In this project we create a hub for each endpoint
	// There is another option, we can create a single instance of hub and
	// and all endpoints (if we have multiple endpoints) use that centeralized hub
	// In that situation we must
	hub      *Hub      // Each endpoint must assiciated with a hub to manage the connections
	stopChan chan bool // notify all goroutines to shutdown

}

// NewEndpoint creates a endpoint for handle configurations
func NewEndpoint(config EndpointConfing) *Endpoint {
	return &Endpoint{
		config:   config,
		stopChan: make(chan bool),
		hub:      NewHub(10),
	}
}

//TODO: conn.SetReadDeadline
//TODO: func (c *TCPConn) SetKeepAlive(keepalive bool) os.Error

// Start listening to the port and reporting new connection
func (e *Endpoint) Start() error {
	addr, errAddr := net.ResolveTCPAddr(e.config.NetType, e.config.GetHostAddress())
	if errAddr != nil {
		log.Printf("Endpoint, Address is not valid %s\n. Error message %s",
			e.config.GetHostAddress(), errAddr.Error())
		return errAddr
	}

	listener, errListen := net.ListenTCP(e.config.NetType, addr)
	if errListen != nil {
		log.Printf("Endpoint, Unable to listen on host address %s\n. Error message %s",
			e.config.GetHostAddress(), errListen.Error())
		return errListen
	}

	//TODO: may be required to close the hub and all connection if server not listening to the port again
	defer listener.Close()
	e.listener = listener

	log.Println("Endpoint, Listen on", listener.Addr().String())
	for {
		select {
		case <-e.stopChan:
			log.Println("Endpoint, exist command recieved")
			return nil
		default:
		}
		log.Println("Endpoint, Accept a connection request.")

		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Endpoint, Failed accepting a connection request:", err)
			continue

		}

		//On OSX and SetKeepAlive this will cause up to 8 TCP keepalive probes to be sent at an
		//interval of 75 seconds after a connection has been idle for 2 hours.
		//Or in other words, Read will return an io.EOF error after 2 hours and 10 minutes (7200 + 8 * 75)
		conn.SetKeepAlive(true)
		//rand.Uint.Seed(time.Now().UTC().UnixNano())
		skt := socket.NewTCPSocket(conn, rand.Uint64(), e.config.SendQueueSize, e.config.ReadBufSize, e.config.WriteBufSize)
		e.hub.Add(skt)
		// ass connection to channel
	}
}

// Stop listening to the port and reporting new connection
func (e *Endpoint) Stop() error {
	err := (e.listener).Close()
	if err != nil {
		log.Printf("Endpoint, Unable to stop host address %s\n. Error message %s",
			e.config.GetHostAddress(), err.Error())
		return err
	}
	e.stopChan <- true
	return nil
}
