package main

import (
	"fmt"
	. "chord"
	"flag"
	"strings"
	"net/rpc"
	"net"
	"net/http"
	"log"
	"os"
)

func main() {
	fsPathPtr := flag.String("fs", "./", "Path of directory to save files")
	neighbourIPPtr := flag.String("n", "", "IP of a neighbour")
	nodeIPPtr := flag.String("ip", "", "IP of the node")
	truncationBits := flag.Int("m", 7, "Size the hash is to be truncated to")
	flag.Parse()

	if strings.Compare(*nodeIPPtr, "") == 0 {
		fmt.Println("IP should be provided")
		os.Exit(1)
	}

	localNode := CreateNode(*nodeIPPtr, *truncationBits)
	localNode.LoadFiles(*fsPathPtr, *truncationBits)

	ring := CreateRing(localNode, *truncationBits)
	localRNode := new(RemoteNode)
	localRNode.Node = localNode
	localRNode.Ring = ring
	rpc.Register(localRNode)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	if strings.Compare(*neighbourIPPtr, "") != 0 {
		client, err := rpc.DialHTTP("tcp", *neighbourIPPtr + ":1234")
		if err != nil {
			log.Fatal("dialing:", err)
		}
		client.Go("RemoteNode.Ring.GetNodes", " ", &ring.Nodes, nil)
	}

	for {
		var cmd, args, output string
		fmt.Scanln(&cmd)
		fmt.Scanln(&args)
		switch cmd {
		case "get":
			key := GetHash(args, *truncationBits)
			ring.Search(key, &output)
		case "put":
			arg := strings.Split(args, ":")
			// fmt
			filePath := arg[0]
			fileName := arg[1]
			ring.AddFile(filePath, fileName)
		default:
			fmt.Println("Unknown command")
		}
	}
}