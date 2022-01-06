package config

import "fmt"

type ErrParamNameDuplicate struct {
	ProxyName string
	Parameter Param
}

// Error returns the string representation of ErrParamNameDuplicate
func (e ErrParamNameDuplicate) Error() string {
	return fmt.Sprintf("config: proxy '%s' redeclares duplicate parameter with name '%s'",
		e.ProxyName, e.Parameter.Name)
}
