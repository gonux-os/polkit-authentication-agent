package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/godbus/dbus/v5"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
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

type AuthenticationRequest struct {
	agent      *Agent
	cookie     string
	ActionId   string
	Message    string
	IconName   string
	Details    map[string]string
	Identities []PKIdentity
	wasDenied  bool
}

func (req *AuthenticationRequest) Accept(identity PKIdentity) error {
	err := req.agent.call("AuthenticationAgentResponse", req.cookie, identity) // Confirm
	if err != nil {
		return fmt.Errorf("sending response: %w", err)
	}
	return nil
}

func (req *AuthenticationRequest) Deny() {
	req.wasDenied = true
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

func run(req *AuthenticationRequest, w *app.Window, message string) error {
	th := material.NewTheme()
	var ops op.Ops
	var acceptButton widget.Clickable
	var denyButton widget.Clickable
	var passwordInput widget.Editor

	go func() {
		for {
			if acceptButton.Clicked() {
				err := req.Accept(req.Identities[0])
				if err != nil {
					panic(err)
				}
				w.Perform(system.ActionClose)
			}
			if denyButton.Clicked() {
				req.Deny()
				w.Perform(system.ActionClose)
			}
		}
	}()

	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			layout.Flex{
				// Vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceAround,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						title := material.Body1(th, message)
						black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
						title.Color = black
						title.Alignment = text.Middle
						return title.Layout(gtx)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(th, &passwordInput, "Password")
						return ed.Layout(gtx)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							// Vertical alignment, from top to bottom
							Axis: layout.Horizontal,
							// Empty space is left at the start, i.e. at the top
							Spacing: layout.SpaceBetween,
						}.Layout(gtx,
							layout.Rigid(
								func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(th, &acceptButton, "Accept")
									return btn.Layout(gtx)
								},
							),
							layout.Rigid(
								func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(th, &denyButton, "Deny")
									return btn.Layout(gtx)
								},
							),
						)
					},
				),
			)

			e.Frame(gtx.Ops)
		}
	}
	return nil
}

func (a *Agent) BeginAuthentication(action_id string, message string, icon_name string, details map[string]string, cookie string, identities []PKIdentity) *dbus.Error {
	fmt.Printf("Auth request > id: %v ; message: %v ; icon_name: %v ; details: %v ; cookie: %v ; identities: %v\n", action_id, message, icon_name, details, cookie, identities)

	fmt.Println(message)

	w := app.NewWindow()
	req := &AuthenticationRequest{
		agent:      a,
		ActionId:   action_id,
		Message:    message,
		IconName:   icon_name,
		Details:    details,
		cookie:     cookie,
		Identities: identities,
	}
	err := run(req, w, message)
	if err != nil {
		panic(fmt.Errorf("executing GUI: %w", err))
	}

	if req.wasDenied {
		return dbus.NewError("org.freedesktop.PolicyKit1.Error.Cancelled", nil)
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

	app.Main()
}
