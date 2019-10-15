package types

type Host struct {
	User        string `json:"user"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	KeyPath     string `json:"keypath"`
	Description string `json:"description"`
}

// Check has password or pem path.
func (h *Host) HasCredentials() bool {
	if h.Password == "" && h.KeyPath == "" {
		return false
	}
	return true
}
