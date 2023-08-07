package gos7_test

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/robinson/gos7"
)

const (
	tcpDevice = "192.168.72.129"
	rack      = 0
	slot      = 2
)

// ClientTestAll client test all
func ClientTestAll(t *testing.T, client gos7.Client) {
	//write value to 100
	ClientTestWriteIntDB(t, client, 100)
	//read and assert with 100
	ClientTestReadIntDB(t, client)
	//return 0
	ClientTestWriteIntDB(t, client, 0)
	//test directory
	ClientTestDirectory(t, client)
	//Get CPU info
	ClientTestGetCPUInfo(t, client)
	//Get AG Block Info
	ClientTestGetAGBlockInfo(t, client)
	//get PLC status
	ClientPLCGetStatus(t, client)
	//multi write to DB2710 -> 1, DB2810 ->2
	ClientAGWriteMulti(t, client)
	//multi read
	ClientAGReadMulti(t, client)
}

// ClientTestWriteIntDB client test write int
func ClientTestWriteIntDB(t *testing.T, client gos7.Client, value int16) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	address := 2710
	start := 8
	size := 2
	buffer := make([]byte, 255)

	//binary.BigEndian.PutUint16(buffer[0:], uint16(value))
	var helper gos7.Helper
	helper.SetValueAt(buffer, 0, value)
	err := client.AGWriteDB(address, start, size, buffer)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, err, nil) // send success then the response in position 6 will be 128
}

// ClientTestReadIntDB client test read int
func ClientTestReadIntDB(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	address := 2710
	start := 8
	size := 2
	buf := make([]byte, 255)
	err := client.AGReadDB(address, start, size, buf)
	if err != nil {
		t.Fatal(err)
	}
	// result := binary.BigEndian.Uint16(results)
	var s7 gos7.Helper
	var result uint16
	s7.GetValueAt(buf, 0, &result)
	AssertEquals(t, 100, int(result))
}

// ClientTestDirectory test directory functions, list all blocks
func ClientTestDirectory(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	//var bl gos7.S7BlocksList
	bl, err := client.PGListBlocks()
	if err != nil {
		t.Fatal(err)
	}
	//for example
	AssertEquals(t, len(bl.OBList), 10)
	AssertEquals(t, len(bl.DBList), 113)
	AssertEquals(t, len(bl.FBList), 81)
}

// ClientTestGetCPUInfo get the CPU info
func ClientTestGetCPUInfo(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	info, err := client.GetCPUInfo()
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, strings.Index(info.SerialNumber, "0118701484"), 0) //return serial should be "0118701484        ", some space
}

// ClientTestGetAGBlockInfo get AG block info
func ClientTestGetAGBlockInfo(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	info, err := client.GetAgBlockInfo(65, 2710)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, info.CodeDate, "22.01.2018")
}

// ClientPLCGetStatus get PLC status
func ClientPLCGetStatus(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	status, err := client.PLCGetStatus()
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, status, 8) //8=running, 4=stop, 0=unknown
}

// ClientAGReadMulti read multi client
func ClientAGReadMulti(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	data1 := make([]byte, 1024)
	data2 := make([]byte, 1024)
	data3 := make([]byte, 1024)
	var error1, error2, error3 string

	var items = []gos7.S7DataItem{
		{
			Area:     0x84,
			WordLen:  0x02,
			DBNumber: 2710,
			Start:    0,
			Amount:   16,
			Data:     data1,
			Error:    error1,
		},
		{
			Area:     0x84,
			WordLen:  0x02,
			DBNumber: 2810,
			Start:    0,
			Amount:   16,
			Data:     data2,
			Error:    error2,
		},
		{
			Area:     0x84,
			WordLen:  0x02,
			DBNumber: 2910,
			Start:    0,
			Amount:   16,
			Data:     data3,
			Error:    error3,
		},
	}
	err := client.AGReadMulti(items, 3)
	if err != nil {
		t.Fatal(err)
	}
	value1 := binary.BigEndian.Uint16(data1[8:]) //in ClientAGWriteMulti wrote all to 1, then output should be 256 + 1
	value2 := binary.BigEndian.Uint16(data2[8:]) //
	value3 := binary.BigEndian.Uint16(data3[8:]) //

	AssertEquals(t, value1, uint16(257))
	AssertEquals(t, value2, uint16(514))
	AssertEquals(t, value3, uint16(0))
}

// ClientAGWriteMulti read multi client
func ClientAGWriteMulti(t *testing.T, client gos7.Client) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	data1 := make([]byte, 1024)
	data2 := make([]byte, 1024)
	data3 := make([]byte, 1024)

	for i := 0; i < 16; i++ {
		data1[i] = 0x01
		data2[i] = 0x02
	}
	var error1, error2, error3 string

	var items = []gos7.S7DataItem{
		{
			Area:     0x84,
			WordLen:  0x02,
			DBNumber: 2710,
			Start:    0,
			Amount:   16,
			Data:     data1,
			Error:    error1,
		},
		{
			Area:     0x84,
			WordLen:  0x02,
			DBNumber: 2810,
			Start:    0,
			Amount:   16,
			Data:     data2,
			Error:    error2,
		},
		{
			Area:     0x84,
			WordLen:  0x02,
			DBNumber: 2910,
			Start:    0,
			Amount:   16,
			Data:     data3,
			Error:    error3,
		},
	}
	err := client.AGWriteMulti(items, 3)
	if err != nil {
		t.Fatal(err)
	}
	if error1 != "" || error2 != "" || error3 != "" {
		t.Fatal(error1 + error2 + error3)
	}
	//value1 := binary.BigEndian.Uint16(data1[8:])
	AssertEquals(t, "", error1)
}

// AssertEquals helper
func AssertEquals(t *testing.T, expected, actual interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	} else {
		// Get file name only
		idx := strings.LastIndex(file, "/")
		if idx >= 0 {
			file = file[idx+1:]
		}
	}

	if expected != actual {
		t.Logf("%s:%d: Expected: %+v (%T), actual: %+v (%T)", file, line,
			expected, expected, actual, actual)
		t.FailNow()
	}
}

func TestTCPClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 200 * time.Second
	handler.IdleTimeout = 200 * time.Second
	handler.Logger = log.New(os.Stdout, "tcp: ", log.LstdFlags)
	handler.Connect()
	defer handler.Close()
	client := gos7.NewClient(handler)
	ClientTestAll(t, client)
}

func TestMultiTCPClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var handlers sync.Map
	var clients sync.Map

	tcpDevices := make([]map[string]string, 2)
	tcpDevices[0] = make(map[string]string, 1)
	tcpDevices[1] = make(map[string]string, 1)
	tcpDevices[0]["tcpDevice"] = "192.168.10.19:102"
	tcpDevices[1]["tcpDevice"] = "192.168.10.10:102"

	c := make(chan int)

	for k := range tcpDevices {
		go func(device map[string]string) {
			handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
			handler.Timeout = 200 * time.Second
			handler.IdleTimeout = 200 * time.Second
			handler.Logger = log.New(os.Stdout, "tcp: ", log.LstdFlags)
			handler.Address = device["tcpDevice"]
			handler.Connect()
			handlers.Store(device["tcpDevice"], handler)

			client := gos7.NewClient(handler)
			clients.Store(device["tcpDevice"], client)

			c <- 1
		}(tcpDevices[k])
	}

	var cS []int

	for n := range c {
		cS = append(cS, n)
		if len(cS) == len(tcpDevices) {
			close(c)
			break
		}
	}

	cli, exist := clients.Load("192.168.10.10:102")
	client, ok := cli.(gos7.Client)
	if exist && ok {
		buf := make([]byte, 255)
		client.AGReadDB(200, 34, 4, buf)
		var s7 gos7.Helper
		var result float32
		s7.GetValueAt(buf, 0, &result)
		fmt.Printf("%v\n", result)
	}

	defer func() {
		handlers.Range(func(key, value interface{}) bool {
			h, _ := handlers.Load(key)
			if hh, ok := h.(*gos7.TCPClientHandler); ok {
				hh.Close()
			}
			return true
		})
	}()
}
