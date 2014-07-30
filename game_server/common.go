package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/gogames/go_tetris/utils"
)

var errNilConn = fmt.Errorf("the connection is nil")

// receive data
func recv(conn *net.TCPConn) (d requestData, err error) {
	if conn == nil {
		err = errNilConn
		return
	}
	b, err := utils.ReadDataOverTcp(conn)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &d)
	return
}

// send data
func send(conn *net.TCPConn, desc string, data interface{}) error {
	if conn == nil {
		return errNilConn
	}
	err := utils.SendDataOverTcp(conn, newResponse(desc, data).toJson())
	if err != nil {
		log.Debug("can not send response data ->\ndesc: %v, data: %v, error: %v", desc, data, err)
	}
	return err
}

// close a connection
func closeConn(conns ...*net.TCPConn) {
	for _, conn := range conns {
		if err := conn.Close(); err != nil {
			log.Debug("can not close the connection: %v", err)
		}
	}
}

// response from game server to client
type responseData struct {
	Description string      `json:"desc"`
	Data        interface{} `json:"data"`
}

func newResponse(desc string, data interface{}) responseData {
	return responseData{Description: desc, Data: data}
}

func (r responseData) String() string {
	return fmt.Sprintf("\ndescription: %v, data: %v", r.Description, r.Data)
}

func (r responseData) toJson() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		log.Debug("can not json marshal the data %v: %v", r, err)
	}
	return b
}

// request from client to game server
type requestData struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

func (d requestData) String() string {
	return fmt.Sprintf("\nrequest command: %s, data: %s", d.Cmd, d.Data)
}
