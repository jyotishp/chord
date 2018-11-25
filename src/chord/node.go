package chord

import (
	"io/ioutil"
	"log"
)

// Struct for remote nodes
type RemoteNode struct {
	Node	*Node
	Ring	*Ring
}

// Struct that defines a node on a ring
type Node struct {
	Ip			string
	Id			int
	Successor	*RemoteNode
	FingerTable	[]int
	FileList	map[int]string
	FsPath		string
}

// Create a node
func CreateNode(ip string, m int) *Node {
	node := new(Node)
	node.Ip = ip
	node.Id = GetHash(ip, m)
	node.FingerTable = make([]int,m)
	node.FileList = make(map[int]string)
	return node
}

// Load files from local FS
func (node *Node) LoadFiles(fsPath string, m int) {
	node.FsPath = fsPath
	files, err := ioutil.ReadDir(fsPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		node.FileList[GetHash(file.Name(), m)] = file.Name()
	}
}