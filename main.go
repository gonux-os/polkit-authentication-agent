package main

import (
	"fmt"
	"gonux/polkit-authentication-agent/agent"
	"gonux/polkit-authentication-agent/cli"
	"gonux/polkit-authentication-agent/config"
	"gonux/polkit-authentication-agent/ui"
	"gonux/polkit-authentication-agent/utils"
)

func main() {
	config.LoadConfig()

	a, err := agent.NewAgent()
	if err != nil {
		panic(err)
	}
	defer a.Close()

	a.OnRequest(cli.LogRequest)
	a.OnRequest(ui.Show)

	err = a.Register()
	if err != nil {
		panic(err)
	}

	fmt.Println("Running")

	utils.SetUserID()
	ui.Boot()
}
