package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mesia777/berith-utils/db"
	"github.com/mesia777/berith-utils/types"
	"log"
)

// AddNode save node into database
func AddNode(db *db.Database, node *types.Node) error {
	if !node.HasCredentials() {
		nodeJson := node.Name
		b, err := json.Marshal(node)
		if err == nil {
			nodeJson = string(b)
		}
		return errors.New("must have at least password or key path :" + nodeJson)
	}

	has, err := db.Has(getNodeKey(node.Name))
	if err != nil {
		return err
	}
	if has {
		return errors.New("already exist node " + node.Name)
	}

	key := getNodeKey(node.Name)
	encoded, err := json.Marshal(node)
	if err != nil {
		return err
	}

	err = db.Put(key, encoded)
	if err != nil {
		return err
	}
	log.Println("success to save a host : ", string(encoded))
	return nil
}

// GetNode returns a node from data store given node name
func GetNode(db *db.Database, name string) (*types.Node, error) {
	val, err := db.Get(getNodeKey(name))
	if err != nil {
		return nil, err
	}

	var n *types.Node
	err = json.Unmarshal(val, &n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// GetNodes returns all node from local store
func GetNodes(db *db.Database) ([]*types.Node, error) {
	itr := db.NewIteratorWithPrefix([]byte(types.NodePrefix))
	var nodes []*types.Node

	for itr.Next() {
		var n *types.Node
		if err := json.Unmarshal(itr.Value(), &n); err != nil {
			fmt.Println("failed to unmarshal node", err)
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// UpdateNode update a given node into data store
func UpdateNode(db *db.Database, update *types.Node) error {
	f, err := GetNode(db, update.Name)
	if err != nil {
		return err
	}

	// FIXME : more efficient compare
	if f.Name != update.Name {
		f.Name = update.Name
	}
	if f.Host.Address != update.Host.Address {
		f.Host.Address = update.Host.Address
	}
	if f.Host.Password != update.Host.Password {
		f.Host.Password = update.Host.Password
	}
	if f.Host.Port != update.Host.Port {
		f.Host.Port = update.Host.Port
	}
	if f.Host.User != update.Host.User {
		f.Host.User = update.Host.User
	}
	if f.Host.KeyPath != update.Host.KeyPath {
		f.Host.KeyPath = update.Host.KeyPath
	}
	if f.Host.Description != update.Host.Description {
		f.Host.Description = update.Host.Description
	}

	err = AddNode(db, f)
	if err == nil {
		log.Println("success to update")
	}
	return err
}

// DeleteHost delete a node given node name
func DeleteHost(db *db.Database, name string) error {
	return db.Delete(getNodeKey(name))
}

// getNodeKey returns a key in data store given node name
func getNodeKey(name string) []byte {
	return []byte(types.NodePrefix + name)
}
