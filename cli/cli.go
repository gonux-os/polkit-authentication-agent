package cli

import (
	"fmt"
	"gonux/polkit-authentication-agent/agent"
)

func LogRequest(req *agent.AuthenticationRequest) error {
	fmt.Printf("Auth request: %s\n", req.Message)
	return nil
}
