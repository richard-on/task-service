package request

type AddRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Coordinators []string `json:"coordinators"`
}
