package connection

import (
	"os"
	"strings"
)

const DefaultURI = "qemu:///system"

type Config struct {
	URI      string
	Username string
	Password string
}

func Resolve(explicitURI, explicitUser, explicitPassword string) Config {
	uri := strings.TrimSpace(explicitURI)
	if uri == "" {
		uri = strings.TrimSpace(os.Getenv("LIBVIRT_DEFAULT_URI"))
	}
	if uri == "" {
		uri = DefaultURI
	}

	user := strings.TrimSpace(explicitUser)
	if user == "" {
		user = strings.TrimSpace(os.Getenv("LIBVIRT_USERNAME"))
	}

	password := explicitPassword
	if password == "" {
		password = os.Getenv("LIBVIRT_PASSWORD")
	}

	return Config{
		URI:      uri,
		Username: user,
		Password: password,
	}
}
