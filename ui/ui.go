package ui

import (
	"gonux/polkit-authentication-agent/agent"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var a fyne.App
var w fyne.Window

func Boot() {
	a = app.New()
	w = a.NewWindow("Authentication required")
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	a.Run()
}

func Show(req *agent.AuthenticationRequest) error {

	closeWindow := make(chan bool)
	defer close(closeWindow)

	w.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			widget.NewLabel(req.Message),
			container.New(
				layout.NewHBoxLayout(),
				widget.NewButton("Deny", func() {
					req.Deny()
					closeWindow <- true
				}),
				widget.NewButton("Accept", func() {
					req.Accept(req.Identities[0])
					closeWindow <- true
				}),
			),
		),
	)
	w.Show()

	for c := range closeWindow {
		if c {
			break
		}
		time.Sleep(1 * time.Nanosecond)
	}

	w.Hide()
	return nil
}
