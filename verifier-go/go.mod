module github.com/moda-gov-tw/twdiw-verifier-go

go 1.22.3

replace github.com/moda-gov-tw/twdiw-issuer-go => ./issuer-go

require github.com/moda-gov-tw/twdiw-issuer-go v0.0.0-00010101000000-000000000000

require github.com/golang-jwt/jwt/v5 v5.3.0

require (
	github.com/fxamacker/cbor/v2 v2.5.0 // indirect
	github.com/veraison/go-cose v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
)
