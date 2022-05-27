package param

// The value of a parameter must be JSON serializable.
type Param struct {
	Val string `json:"val"`
}
