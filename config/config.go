package config

type operationMode int64

const (
	_               operationMode = iota
	AcceptAll                     // Automatically accept all requests
	DenyAll                       // Automatically deny all requests
	RequestYesNo                  // Request confirmation from the user
	RequestPassword               // Request confirmation from the user with password validation
)

type config struct {
	Mode               operationMode
	NotifyOnAutoAccept bool
	NotifyOnAutoDeny   bool
	ShowUsername       bool
}

var Global config

func LoadConfig() {
	Global = config{
		Mode:               RequestPassword,
		NotifyOnAutoAccept: true,
		NotifyOnAutoDeny:   false,
		ShowUsername:       true,
	}
}
