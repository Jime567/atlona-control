package actions

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"time"
)

// Creating Connection
func createConnection(address string, port string) (*net.TCPConn, error) {
	radder, err := net.ResolveTCPAddr("tcp", address+":"+port)
	if err != nil {
		err = fmt.Errorf("error resolving address : %s", err.Error())
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, radder)
	if err != nil {
		err = fmt.Errorf("error dialing address : %s", err.Error())
		return nil, err
	}

	return conn, nil
}

// SendCommand opens a connection with <addr> and sends the <command> to the via, returning the response, or an error if one occured.
func sendCommand(address string, port string, cmd []byte) ([]byte, error) {
	fmt.Printf("\n\nOpening telnet connection with address: %s command decimal: %d, command: %s\n", address, cmd, string(cmd))
	conn, err := createConnection(address, port)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	timeoutDuration := 100 * time.Millisecond

	conn.SetReadDeadline(time.Now().Add(timeoutDuration))

	reader := bufio.NewReader(conn)
	resp := make([]byte, 64)
	readTime := time.Millisecond * 100
	for start := time.Now(); ; {
		_, err := reader.Read(resp)
		time.Sleep(time.Duration(5 * time.Millisecond))
		if err != nil {
			break
		}

		fmt.Printf("The initial response is: %s\r\n", resp)
		if strings.Contains(string(resp), "Full Connections") {
			err := fmt.Errorf("error in switcher response:  %s", string(resp))
			conn.Close()
			return nil, err
		}

		if time.Since(start) > readTime {
			break
		}
	}
	if len(resp) < 1 {
		err := fmt.Errorf("no initial response, closing connection")
		conn.Close()
		return nil, err
	}

	fmt.Println("Write Command: ", string(cmd))
	if len(cmd) > 0 {
		_, err = conn.Write(cmd)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}
	resp = nil //clear out the connection initialized read

	//read response to command - may get multiple responses (aka multiple \n's)
	timeoutDuration = 500 * time.Millisecond
	conn.SetReadDeadline(time.Now().Add(timeoutDuration))
	readTime = time.Millisecond * 50
	for start := time.Now(); ; {
		if time.Since(start) > readTime {
			break
		}
		tempResp, err := reader.ReadBytes('\n')
		if len(tempResp) > 0 {
			resp = append(resp, tempResp...)
		}
		if err != nil {
			continue
		}
	}
	fmt.Printf("The second response is: %s\r\n", resp)
	fmt.Println("decimal: ", resp)
	if err != nil {
		conn.Close()
		return nil, err
	}

	//catch for failed command response from Atlona
	if string(resp) == "Command FAILED" {
		err = fmt.Errorf("failed command, please check the correct device is selected. device response: %s", string(resp))
		conn.Close()
		return nil, err
	}

	conn.Close()
	//replace all commas with \r\n in response
	resp = []byte(strings.Replace(string(resp), ",", "\r\n", -1))
	//replace all Uppercase X's with lowercase x's
	resp = []byte(strings.Replace(string(resp), "X", "x", -1))
	return resp, nil
}
