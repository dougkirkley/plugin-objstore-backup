package provider

type Provider string

const (
	S3 Provider = "s3"
)

func (p Provider) String() string {
	return string(p)
}

var (
	providers = []Provider{S3}
)

func Validate(provider string) bool {
	for _, p := range providers {
		if provider == string(p) {
			return true
		}
	}
	return false
}
