package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Anbcorp/sa3toNMEA/runners"
	nmea "github.com/Anbcorp/sa_nmea"
	"github.com/tidwall/geodesic"
)

type client struct {
	conn net.Conn
}

func main() {
	b := nmea.NMEABoat{
		Hdg: 279,
		Spd: 28.2,
		Cog: 281,
		Sog: 28.2,
		Tws: 13.7,
		//Twd:          238,
		Twa:       -41,
		Awa:       -26,
		Aws:       20.7,
		Latitude:  46.5086619986689,
		Longitude: -9.57924805707325,
		Timestamp: time.Now().UTC(),
	}

	upd := runners.NewUpdater(6, &b, new(sync.RWMutex))

	go upd.Run()

	viewer := runners.NewTask(1)

	go func() {
		for {
			select {
			case <-viewer.Done:
				viewer.Ticker.Stop()
				return
			case <-viewer.Ticker.C:
				fmt.Println(time.Now(), upd.Boat.Cog)
			}
		}
	}()

	time.Sleep(120 * time.Second)
	upd.Stop()
	viewer.Done <- true
}

func realmain() {
	b := nmea.NMEABoat{
		Hdg: 279,
		Spd: 28.2,
		Cog: 281,
		Sog: 28.2,
		Tws: 13.7,
		//Twd:          238,
		Twa:       -41,
		Awa:       -26,
		Aws:       20.7,
		Latitude:  46.5086619986689,
		Longitude: -9.57924805707325,
		Timestamp: time.Now().UTC(),
	}

	list := []string{"GLL", "GGA", "VHW", "HDT", "MWV", "MWV.R", "VTG", "RMC"}
	var newLat, newLon float64

	geodesic.WGS84.Direct(b.Latitude, b.Longitude, b.Cog, b.Sog/360, &newLat, &newLon, nil)

	// Define the port to listen on
	port := ":10112"

	// Start listening for incoming connections
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server is listening on %s\n", port)

	// A map to store active clients
	clients := make(map[net.Conn]*client)
	var mu sync.Mutex

	// Goroutine to handle broadcasting messages to clients every 10 seconds
	go func() {
		for {
			log.Printf("Sleeping...\n")
			time.Sleep(1 * time.Second)
			// Update boat position
			geodesic.WGS84.Direct(b.Latitude, b.Longitude, b.Cog, b.Sog/360, &newLat, &newLon, nil)
			b.Latitude = newLat
			b.Longitude = newLon
			log.Printf("%f : %f", b.Latitude, b.Longitude)
			mu.Lock()
			for conn := range clients {
				// Send a message to the client
				// message := fmt.Sprintf("Server Time: %s\n", time.Now().Format(time.RFC1123))
				message := nmea.WriteMessage(b, list)
				writer := bufio.NewWriter(conn)
				_, err := writer.WriteString(message)
				if err != nil {
					log.Printf("Error writing to client: %v\n", err)
					conn.Close()
					delete(clients, conn)
					continue
				}
				writer.Flush()
				log.Printf("Message sent to %s", conn.RemoteAddr())
			}
			mu.Unlock()
		}
	}()

	for {
		// Accept a new client connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		log.Printf("New client connected: %s\n", conn.RemoteAddr())

		mu.Lock()
		clients[conn] = &client{conn: conn}
		mu.Unlock()

		// Goroutine to handle client-specific activity
		go func(conn net.Conn) {
			defer func() {
				mu.Lock()
				delete(clients, conn)
				mu.Unlock()
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
				// mu.Lock()
				// clients[conn].lastPing = time.Now()
				// mu.Unlock()
			}
		}(conn)
	}
}
