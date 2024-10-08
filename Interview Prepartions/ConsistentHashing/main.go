package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

type SortedKeys []uint32

func (x SortedKeys) Len() int           { return len(x) }
func (x SortedKeys) Less(i, j int) bool { return x[i] < x[j] }
func (x SortedKeys) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type ConsistentHashing struct {
	// It is number of virtual node. The Virtual node is concept to duplicate node to archieve balance in ConsistentHashing
	// Default number of virtual node is 20, it will be more balance if your incrase this value
	NumOfVirtualNode int

	hashSortedKeys SortedKeys

	circleRing map[uint32]string
	dataSet    map[string]bool
}

// New a Consistent Hashing with default 20 virtual node
func NewConsistentHashing() *ConsistentHashing {
	return &ConsistentHashing{
		NumOfVirtualNode: 20,
		circleRing:       make(map[uint32]string),
		dataSet:          make(map[string]bool),
	}
}

// Get a nearest object name from input object in consistent hashing ring
func (c *ConsistentHashing) Get(obj string) (string, error) {
	if len(c.dataSet) == 0 {
		return "", errors.New("Empty struct")
	}

	nearObj, _ := c.circleRing[c.hashSortedKeys[c.searchNearRingIndex(obj)]]
	return nearObj, nil
}

// Add a node into this consistent hashing ring
func (c *ConsistentHashing) Add(node string) {
	if _, find := c.dataSet[node]; find {
		return
	}

	c.dataSet[node] = true
	key := c.hasKey(node)
	c.circleRing[key] = node

	//Add virtual node for "balance"
	for i := 0; i < c.NumOfVirtualNode; i++ {
		vk := c.getVirtualNodeKey(i, node)
		c.circleRing[vk] = node
	}

	c.updateSortHashKeys()
}

// Remove a node from this consistent hashing ring
func (c *ConsistentHashing) Remove(node string) {
	if _, find := c.dataSet[node]; !find {
		return //not in our dataset
	}

	delete(c.dataSet, node)
	key := c.hasKey(node)
	delete(c.circleRing, key)

	//Delete virtual node
	for i := 0; i < c.NumOfVirtualNode; i++ {
		vk := c.getVirtualNodeKey(i, node)
		delete(c.circleRing, vk)
	}

	c.updateSortHashKeys()
}

// List the whole nodes in consistent hashing ring
func (c *ConsistentHashing) ListNodes() []string {
	var retList []string
	for k, _ := range c.dataSet {
		retList = append(retList, k)
	}
	return retList
}

func (c *ConsistentHashing) getVirtualNodeKey(index int, obj string) uint32 {
	newObjStr := strconv.Itoa(index) + "-" + obj
	return c.hasKey(newObjStr)
}

func (c *ConsistentHashing) searchNearRingIndex(obj string) int {
	targetKey := c.hasKey(obj)

	targetIndex := sort.Search(len(c.hashSortedKeys), func(i int) bool { return c.hashSortedKeys[i] >= targetKey })

	if targetIndex >= len(c.hashSortedKeys) {
		targetIndex = 0
	}

	return targetIndex
}

func (c *ConsistentHashing) updateSortHashKeys() {
	c.hashSortedKeys = nil

	for node, _ := range c.dataSet {
		key := c.hasKey(node)
		c.hashSortedKeys = append(c.hashSortedKeys, key)
	}
	sort.Sort(c.hashSortedKeys)
}

func (c *ConsistentHashing) hasKey(obj string) uint32 {
	var scratch [64]byte
	if len(obj) < 64 {

		copy(scratch[:], obj)
	}
	return crc32.ChecksumIEEE(scratch[:len(obj)])
}


func main() {
	// Create a new consistent hashing instance
	ch := NewConsistentHashing()

	// Add nodes to the consistent hashing ring
	ch.Add("NodeA")
	ch.Add("NodeB")
	ch.Add("NodeC")

	// List and display all nodes
	nodes := ch.ListNodes()
	fmt.Println("Nodes in the ring:")
	for _, node := range nodes {
		fmt.Println(" -", node)
	}

	// Define a set of test keys
	keys := []string{"key1", "key2", "key3", "key4", "key5"}

	// Retrieve and display the assigned node for each key
	for _, key := range keys {
		node, err := ch.Get(key)
		if err != nil {
			log.Fatalf("Error retrieving node for key %s: %v", key, err)
		}
		fmt.Printf("Key '%s' is mapped to node '%s'\n", key, node)
	}

	// Remove a node from the ring
	ch.Remove("NodeB")

	// List and display all nodes after removal
	nodes = ch.ListNodes()
	fmt.Println("\nNodes in the ring after removing NodeB:")
	for _, node := range nodes {
		fmt.Println(" -", node)
	}

	// Retrieve and display the assigned node for each key after removal
	for _, key := range keys {
		node, err := ch.Get(key)
		if err != nil {
			log.Fatalf("Error retrieving node for key %s: %v", key, err)
		}
		fmt.Printf("Key '%s' is now mapped to node '%s'\n", key, node)
	}
}
