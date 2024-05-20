package data

import "encoding/base64"

func (i Image) Base64() string {
	return base64.StdEncoding.EncodeToString(i.Data)
}

func (i Image) MediaType() string {
	if i.Extension == "png" {
		return "image/png"
	}
	if i.Extension == "jpg" {
		return "image/jpeg"
	}
	return ""
}

func (i Image) DataURI() string {
	if i.Extension == "png" {
		return "data:image/png;base64," + i.Base64()
	}
	if i.Extension == "jpg" {
		return "data:image/jpeg;base64," + i.Base64()
	}
	if i.Extension == "jpeg" {
		return "data:image/jpeg;base64," + i.Base64()
	}
	return ""
}
