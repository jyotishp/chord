package chord

import (
	"math"
	"log"
	"io/ioutil"
	"fmt"
	"net/rpc"
)

// Struct for Ring object
type Ring struct {
	TruncationBit	int
	Nodes			map[int]*RemoteNode
	LocalNode		*Node
}

// Create a node
func CreateRing(localNode *Node, truncationBit int) *Ring {
	ring := new(Ring)
	ring.TruncationBit = truncationBit
	ring.Nodes = make(map[int]*RemoteNode)
	ring.LocalNode = localNode
	return ring
}

// Join a node to ring -- Triggered by RPC
func (ring *Ring) JoinNode(node *RemoteNode, output *string) error {
	ring.Nodes[node.Node.Id] = node
	ring.Update()
	return nil
}

// Get node list -- Triggered by RPC
func (ring *Ring) GetNodes(args string, nodes *map[int]*RemoteNode) error {
	*nodes = ring.Nodes
	return nil
}

// Delete a node -- Triggered by RPC
func (ring *Ring) RemoveNode(node *RemoteNode, output *string) error {
	delete(ring.Nodes, node.Node.Id)
	ring.Update()
	return nil
}

// Get the number of nodes in the ring
func (ring *Ring) nodeCount() int {
	return len(ring.Nodes)
}

// Find successor of a node
func (ring *Ring) Successor() {
	// Can be improved using binary search
	for id, node := range ring.Nodes {
		if id > ring.LocalNode.Id {
			ring.LocalNode.Successor = node
			break
		}
	}
}

// Compute finger table of the local node
func (ring *Ring) GetFingerTable() {
	ring.LocalNode.FingerTable = make([]int, 0)
	mod := int(math.Exp2(float64(ring.TruncationBit)))
	for i := 0; i < ring.TruncationBit; i++ {
		value := (ring.LocalNode.Id + int(math.Exp2(float64(i)))) % mod
		ring.LocalNode.FingerTable = append(ring.LocalNode.FingerTable, value)
	}
}

// Update local node parameters
func (ring *Ring) Update() {
	ring.Successor()
	ring.GetFingerTable()
}

// Search for key -- Triggered by RPC
func (ring *Ring) Search(key int, output *string) error {
	*output = ""
	// Perform local search
	fileName, prs := ring.LocalNode.FileList[key]
	if prs {
		data, err := ioutil.ReadFile(ring.LocalNode.FsPath + fileName)
		if err != nil {
			log.Fatal(err)
		}
		*output = string(data)
		return nil
	}
	// Send the request to next node
	for i := len(ring.LocalNode.FingerTable) - 1; i >= 0; i-- {
		if ring.LocalNode.FingerTable[i] <= key {
			nodeID := ring.LocalNode.FingerTable[i]
			client, err := rpc.DialHTTP("tcp", ring.Nodes[nodeID].Node.Ip + ":1234")
			if err != nil {
				log.Fatal("dialing:", err)
			}
			client.Go("RemoteNode.Ring.Search", key, output, nil)
			// ring.nodes[ring.localNode.fingerTable[i]].ring.Search(key, output)
			return nil
		}
	}
	// Send the key to successor in case of inconsistencies
	ring.LocalNode.Successor.Ring.Search(key, output)
	return nil
}

// Insert file on a node -- Triggered by RPC
func (ring *Ring) InsertFileByRPC(file *File, output *bool) error {
	err := ioutil.WriteFile(ring.LocalNode.FsPath + file.Name, file.Data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	ring.LocalNode.FileList[GetHash(file.Name, ring.TruncationBit)] = file.Name
	return nil
}

// Find insertion point and insert to node
func (ring *Ring) AddFile(filePath, fileName string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	file := new(File)
	file.Name = fileName
	file.Data = data
	key := GetHash(fileName, ring.TruncationBit)

	for id, node := range ring.Nodes {
		if id >= key {
			if id == ring.LocalNode.Id {
				err := ioutil.WriteFile(ring.LocalNode.FsPath + file.Name, file.Data, 0644)
				if err != nil {
					log.Fatal(err)
				}
				ring.LocalNode.FileList[key] = file.Name
			} else {
				var tmp bool
				client, err := rpc.DialHTTP("tcp", node.Node.Ip + ":1234")
				if err != nil {
					log.Fatal("dialing:", err)
				}
				client.Go("RemoteNode.Ring.InsertFileByRPC", file, &tmp, nil)
				// ring.nodes[id].ring.InsertFileByRPC(file, &_)
			}
		return
		}
	}

	// fmt.Println(ring.LocalNode.FsPath)
	err = ioutil.WriteFile(ring.LocalNode.FsPath + file.Name, file.Data, 0644)
	// fmt.Println("Test2")
	if err != nil {
		log.Fatal(err)
	}
	ring.LocalNode.FileList[key] = file.Name
}