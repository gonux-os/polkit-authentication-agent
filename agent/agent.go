package agent

import (
	"fmt"
	"os"
	"os/user"
	"syscall"

	"github.com/godbus/dbus/v5"
)

type PKSubject struct {
	Kind    string                  `dbus:"subject_kind"`
	Details map[string]dbus.Variant `dbus:"subject_details"`
}

type Agent struct {
	conn      *dbus.Conn
	object    dbus.BusObject
	subject   PKSubject
	callbacks []func(req *AuthenticationRequest) error
}

type PKIdentity struct {
	Kind    string                  `dbus:"identity_kind"`
	Details map[string]dbus.Variant `dbus:"identity_details"`
}

type AuthenticationRequest struct {
	agent            *Agent
	cookie           string
	ActionId         string
	Message          string
	IconName         string
	Details          map[string]string
	Identities       []PKIdentity
	wasAccepted      bool
	acceptedIdentity PKIdentity
}

type User struct {
	Uid      uint32
	Username string
	Name     string
}

func (identity *PKIdentity) User() User {
	if identity.Kind != "unix-user" {
		panic("identity kind is not 'unix-user'")
	}
	uid := identity.Details["uid"].Value().(uint32)
	u, err := user.LookupId(fmt.Sprintf("%d", uid))
	if err != nil {
		panic(err)
	}
	return User{
		Uid:      uid,
		Username: u.Username,
		Name:     u.Name,
	}
}

func NewAgent() (*Agent, error) {
	syscall.Seteuid(0)
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
	req := &AuthenticationRequest{
		agent:      a,
		ActionId:   action_id,
		Message:    message,
		IconName:   icon_name,
		Details:    details,
		cookie:     cookie,
		Identities: identities,
	}
	for i, callback := range a.callbacks {
		err := callback(req)
		if err != nil {
			panic(fmt.Errorf("in request callback #%d: %w", i, err))
		}
	}

	if req.wasAccepted {
		err := req.agent.call("AuthenticationAgentResponse", req.cookie, req.acceptedIdentity) // Confirm
		if err != nil {
			panic(fmt.Errorf("sending response: %w", err))
		}
		return nil
	}
	return dbus.NewError("org.freedesktop.PolicyKit1.Error.Cancelled", nil)
}

func (a *Agent) OnRequest(callback func(req *AuthenticationRequest) error) {
	a.callbacks = append(a.callbacks, callback)
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

func (req *AuthenticationRequest) Accept(identity PKIdentity) {
	req.wasAccepted = true
	req.acceptedIdentity = identity
}

func (req *AuthenticationRequest) Deny() {
	req.wasAccepted = false
}
