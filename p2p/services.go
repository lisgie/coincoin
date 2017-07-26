package p2p

import "time"

//this is not accessed concurrently, one single goroutine
func broadcastService() {
	logger.Print("Start broadcasting service.")
	for {
		select {
		//broadcasting all messages
		case msg := <-brdcstMsg:
			for p := range peers {
				p.ch <- msg
			}
		case p := <-register:
			peers[p] = true
		case p := <-disconnect:
			delete(peers, p)
			close(p.ch)
		}
	}
}

//Belongs to the broadcast service
func peerWriter(p *peer) {
	for msg := range p.ch {
		logger.Printf("Sending payload to %v\n", p.conn.RemoteAddr().String())
		sendData(p, msg)
	}
}

//Single goroutine that makes sure the system is well connected
func checkHealthService() {

	for {
		//time.Sleep(time.Minute)
		time.Sleep(10*time.Second)

		if len(peers) >= MIN_MINERS {
			continue
		}

		select {
		case ipaddr := <-iplistChan:
			logger.Printf("New IP rcvd through channel: %v\n", ipaddr)
			p, err := initiateNewMinerConnection(ipaddr)
			if err != nil {
				logger.Printf("Failed to initiate connection with IP address: %v\n", ipaddr)
				continue
			}
			go minerConn(p)
		default:
			neighborReq()
		}
	}
}
