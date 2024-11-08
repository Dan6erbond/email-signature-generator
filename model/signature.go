package model

import "github.com/a-h/templ"

type Signature struct {
	Name        string
	Role        string
	Email       string
	PhoneNumber string
	Picture     string
	LinkedInURL string
	Company     Company
	BrandColor  string
	Attrs       templ.Attributes
}
