package types

var NodePrefix = "node."

type Node struct {
	Name string `json:"name"`
	Host *Host  `json:"host"`
}

// HasCredentials checks has password or pem path or not
func (n *Node) HasCredentials() bool {
	return n.Host.HasCredentials()
}
