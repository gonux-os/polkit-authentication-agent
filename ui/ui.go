package ui

import (
	"gonux/polkit-authentication-agent/agent"
	"gonux/polkit-authentication-agent/config"
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
	a = app.NewWithID("gonux/polkit-authentication-agent")
	w = a.NewWindow("Authentication required")
	a.Run()
}

func Show(req *agent.AuthenticationRequest) error {

	if config.Global.Mode == config.AcceptAll {
		req.Accept(req.Identities[0])
		if config.Global.NotifyOnAutoAccept {
			a.SendNotification(fyne.NewNotification("Request auto-accepted", req.Message))
		}
		return nil
	}

	if config.Global.Mode == config.DenyAll {
		req.Deny()
		if config.Global.NotifyOnAutoDeny {
			a.SendNotification(fyne.NewNotification("Request auto-denied", req.Message))
		}
		return nil
	}

	closeWindow := make(chan bool)
	defer close(closeWindow)

	passwordInput := widget.NewPasswordEntry()
	passwordInput.SetPlaceHolder("Password")
	if config.Global.Mode != config.RequestPassword {
		passwordInput.Hide()
	}

	var selectedIdentity agent.PKIdentity

	var userNames []string
	for _, identity := range req.Identities {
		switch config.Global.UserSelectorField {
		case config.Username:
			userNames = append(userNames, identity.User().Username)
		case config.Name:
			userNames = append(userNames, identity.User().Name)
		}
	}

	userSelect := widget.NewSelect(userNames, func(s string) {})
	selectedIdentity = req.Identities[0]
	userSelect.SetSelectedIndex(0)
	userSelect.OnChanged = func(s string) {
		selectedIdentity = req.Identities[userSelect.SelectedIndex()]
	}
	if !config.Global.ShowUserSelector {
		userSelect.Hide()
	}

	w.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			widget.NewLabel(req.Message),
			userSelect,
			passwordInput,
			container.New(
				layout.NewHBoxLayout(),
				widget.NewButton("Deny", func() {
					req.Deny()
					closeWindow <- true
				}),
				widget.NewButton("Accept", func() {
					req.Accept(selectedIdentity)
					closeWindow <- true
				}),
			),
		),
	)

	w.SetCloseIntercept(func() {
		req.Deny()
		closeWindow <- true
	})

	w.Resize(fyne.NewSize(300, 150))
	w.SetFixedSize(true)
	w.Show()
	w.CenterOnScreen()
	w.RequestFocus()

	w.Canvas().Focus(passwordInput)

	for c := range closeWindow {
		if c {
			break
		}
		time.Sleep(1 * time.Nanosecond)
	}

	w.Hide()
	return nil
}
