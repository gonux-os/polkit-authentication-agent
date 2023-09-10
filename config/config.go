package config

type operationMode int64

const (
	_               operationMode = iota
	AcceptAll                     // Automatically accept all requests
	DenyAll                       // Automatically deny all requests
	RequestYesNo                  // Request confirmation from the user
	RequestPassword               // Request confirmation from the user with password validation
)

type userSelectorField int64

const (
	_ userSelectorField = iota
	Username
	Name
)

type config struct {
	Mode               operationMode
	NotifyOnAutoAccept bool
	NotifyOnAutoDeny   bool
	ShowUserSelector   bool
	UserSelectorField  userSelectorField
}

var Global config

func LoadConfig() {
	Global = config{
		Mode:               RequestPassword,
		NotifyOnAutoAccept: true,
		NotifyOnAutoDeny:   false,
		ShowUserSelector:   true,
		UserSelectorField:  Name,
	}
}
