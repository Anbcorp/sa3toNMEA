package runners

import (
	"bufio"
	"log"
	"net"
	"sync"

	nmea "github.com/Anbcorp/sa_nmea"
)

type client struct {
	conn net.Conn
}

type Server struct {
	Task
	Boat        *nmea.NMEABoat
	boatMut     *sync.RWMutex
	cliMut      *sync.Mutex
	clients     map[net.Conn]*client
	nmeaPhrases []string
	bindAddr    string
	listener    net.Listener
}

func NewServer(d int, b *nmea.NMEABoat, m *sync.RWMutex) Server {
	var nu Server
	nu.Task = NewTask(d)
	nu.Boat = b
	nu.boatMut = m
	nu.cliMut = new(sync.Mutex)
	nu.clients = make(map[net.Conn]*client)
	nu.nmeaPhrases = []string{"GLL", "GGA", "VHW", "HDT", "MWV", "MWV.R", "VTG", "RMC"}
	nu.bindAddr = ":10112"

	return nu
}

func (srv Server) Update() {
	log.Println("srv.update boat lock")
	srv.boatMut.RLock()
	log.Printf("%f : %f", srv.Boat.Latitude, srv.Boat.Longitude)
	srv.boatMut.RUnlock()
	log.Println("srv update cli lock")
	srv.cliMut.Lock()
	for conn := range srv.clients {
		// Send a message to the client
		// message := fmt.Sprintf("Server Time: %s\n", time.Now().Format(time.RFC1123))
		message := nmea.WriteMessage(*srv.Boat, srv.nmeaPhrases)
		writer := bufio.NewWriter(conn)
		_, err := writer.WriteString(message)
		if err != nil {
			log.Printf("Error writing to client: %v\n", err)
			conn.Close()
			delete(srv.clients, conn)
			continue
		}
		writer.Flush()
		log.Printf("Message sent to %s", conn.RemoteAddr())
	}
	srv.cliMut.Unlock()
}

func (srv Server) Run() {
	for {
		select {
		case <-srv.Done:
			srv.Ticker.Stop()
			srv.listener.Close()
			return
		case <-srv.Ticker.C:
			srv.Update()
		}
	}
}

func (srv Server) Start() {
	var err error
	srv.listener, err = net.Listen("tcp", srv.bindAddr)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.listener.Close()
	log.Printf("Server is listening on %s\n", srv.bindAddr)
	// Start broadcasting
	go srv.Run()

	for {
		// Accept a new client connection
		conn, err := srv.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		log.Printf("New client connected: %s\n", conn.RemoteAddr())
		log.Println(srv.Ticker)

		srv.cliMut.Lock()
		srv.clients[conn] = &client{conn: conn}
		srv.cliMut.Unlock()

		// Goroutine to handle client-specific activity
		go func(conn net.Conn) {
			defer func() {
				srv.cliMut.Lock()
				delete(srv.clients, conn)
				srv.cliMut.Unlock()
				conn.Close()
				log.Printf("Client disconnected: %s\n", conn.RemoteAddr())
			}()

			// Read data to keep track of activity
			reader := bufio.NewReader(conn)
			for {
				_, err := reader.ReadString('\n') // Simulate client activity
				if err != nil {
					return
				}
				log.Printf("Client ping: %s\n", conn.RemoteAddr())
				// mu.Lock()
				// clients[conn].lastPing = time.Now()
				// mu.Unlock()
			}
		}(conn)
	}
}

func (srv Server) Stop() {
	log.Println("Server stopping...")
	srv.Done <- true
}
