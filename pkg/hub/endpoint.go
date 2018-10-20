package hub

import (
	"fmt"
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
	HubQueueSize  int
}

// GetHostAddress Apprend host address and port number together and return  full address of the site
func (conf *EndpointConfing) GetHostAddress() string {
	return conf.Host + ":" + strconv.Itoa(conf.Port)
}

// Endpoint is tcp endpint that handle input connections
type Endpoint struct {
	config   EndpointConfing  // Server configuration
	listener *net.TCPListener // Reference to listener
	// In this project, we create a hub for each endpoint
	// There is another option, we can create a single instance of hub and
	// and all endpoints (if we have multiple endpoints) use that centralized hub
	hub *Hub // Each endpoint must associated with a hub to manage the connections

}

// NewEndpoint creates an endpoint for handle configurations
func NewEndpoint(config EndpointConfing) *Endpoint {
	return &Endpoint{
		config: config,
		hub:    NewHub(config.HubQueueSize),
	}
}

// Start listening to the port and reporting new connection
func (e *Endpoint) Start() error {
	addr, errAddr := net.ResolveTCPAddr(e.config.NetType, e.config.GetHostAddress())
	if errAddr != nil {
		fmt.Printf("Endpoint, Address is not valid %s. Error message %s\n",
			e.config.GetHostAddress(), errAddr.Error())
		return errAddr
	}

	listener, errListen := net.ListenTCP(e.config.NetType, addr)
	if errListen != nil {
		fmt.Printf("Endpoint, Unable to listen on host address %s. Error message %s\n",
			e.config.GetHostAddress(), errListen.Error())
		return errListen
	}

	defer listener.Close()
	e.listener = listener

	fmt.Printf("Endpoint, Listening on %s\n", e.config.GetHostAddress())
	for {

		fmt.Println("Endpoint, Accept a connection request...")
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Endpoint, Failed accepting a connection request. Error message=%s\n", err.Error())
			continue
		}

		//On OSX and SetKeepAlive this will cause up to 8 TCP keepalive probes to be sent at an
		//interval of 75 seconds after a connection has been idle for 2 hours.
		//In other words, Read will return an io.EOF error after 2 hours and 10 minutes (7200 + 8 * 75)
		conn.SetKeepAlive(true)
		skt := socket.NewTCPSocket(conn, rand.Uint64(), e.config.SendQueueSize, e.config.ReadBufSize, e.config.WriteBufSize)
		e.hub.Add(skt)
	}
}
