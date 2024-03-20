package request

type UpdateUserAddressReq struct {
	Address  string `json:"address,omitempty"`
	Location string `json:"location,omitempty"`
}
