package main

import (
	"bytes"
	"text/template"
)

const profileTemplate = `[Interface]
PrivateKey = {{ .PrivateKey }}
Address = {{ .Address1 }}/32
Address = {{ .Address2 }}/128
DNS = 1.1.1.1, 1.0.0.1
[Peer]
PublicKey = {{ .PublicKey }}
AllowedIPs = 0.0.0.0/0
AllowedIPs = ::/0
Endpoint = {{ .Endpoint }}
# device_id = {{ .DeviceID }}
# response = {{ .Response }}`

type ProfileData struct {
	PrivateKey	string
	Address1	string
	Address2	string
	PublicKey	string
	Endpoint	string
	Response	string
	DeviceID	string
}

func GenerateProfile(data *ProfileData) (string, error) {
	t, err := template.New("").Parse(profileTemplate)
	if err != nil {
		return "", err
	}
	var result bytes.Buffer
	if err := t.Execute(&result, data); err != nil {
		return "", err
	}
	return result.String(), nil
}
