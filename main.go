package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/godbus/dbus/v5"
)

type Agent struct {
	conn    *dbus.Conn
	object  dbus.BusObject
	subject PKSubject
}

type PKSubject struct {
	Kind    string                  `dbus:"subject_kind"`
	Details map[string]dbus.Variant `dbus:"subject_details"`
}

type PKIdentity struct {
	Kind    string                  `dbus:"identity_kind"`
	Details map[string]dbus.Variant `dbus:"identity_details"`
}

func NewAgent() (*Agent, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("connecting to DBus: %w", err)
	}
	names := bus.Names()
	if len(names) == 0 {
		panic("empty dbus names")
	}
	sid, exists := os.LookupEnv("XDG_SESSION_ID")
	if !exists {
		sid = "1"
	}
	if err != nil {
		return nil, fmt.Errorf("getting unix session id: %w", err)
	}
	return &Agent{
		conn:   bus,
		object: bus.Object("org.freedesktop.PolicyKit1", "/org/freedesktop/PolicyKit1/Authority"),
		subject: PKSubject{
			Kind: "unix-session",
			Details: map[string]dbus.Variant{
				"session-id": dbus.MakeVariant(sid),
			},
		},
	}, nil
}

func (a *Agent) Close() error {
	return a.conn.Close()
}

func (a *Agent) call(method string, args ...interface{}) error {
	fullMethod := "org.freedesktop.PolicyKit1.Authority." + method
	call := a.object.Call(fullMethod, 0, args...)
	if call.Err != nil {
		return fmt.Errorf("calling method '"+fullMethod+"': %w", call.Err)
	}
	return nil
}

func (a *Agent) BeginAuthentication(action_id string, message string, icon_name string, details map[string]string, cookie string, identities []PKIdentity) *dbus.Error {
	fmt.Printf("Auth request > id: %v ; message: %v ; icon_name: %v ; details: %v ; cookie: %v ; identities: %v\n", action_id, message, icon_name, details, cookie, identities)

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2)

	fmt.Println(style.Render(message))
	// TODO: Show GUI
	err := a.call("AuthenticationAgentResponse", cookie, identities[0]) // Confirm
	if err != nil {
		panic(fmt.Errorf("sending response: %w", err))
	}
	return nil
}

func (a *Agent) Register() error {
	err := a.conn.Export(a, "/org/freedesktop/PolicyKit1/AuthenticationAgent", "org.freedesktop.PolicyKit1.AuthenticationAgent")
	if err != nil {
		return fmt.Errorf("registering listener: %w", err)
	}
	locale := "en_US" // TODO: Fetch from somewhere
	err = a.call("RegisterAuthenticationAgent", a.subject, locale, "/org/freedesktop/PolicyKit1/AuthenticationAgent")
	if err != nil {
		return fmt.Errorf("registering agent: %w", err)
	}
	return nil
}

func main() {
	agent, err := NewAgent()
	if err != nil {
		panic(err)
	}
	defer agent.Close()

	err = agent.Register()
	if err != nil {
		panic(err)
	}

	fmt.Println("Running")

	for {
		time.Sleep(1 * time.Nanosecond)
	}
}
