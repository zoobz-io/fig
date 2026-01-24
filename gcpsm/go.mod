module github.com/zoobzio/fig/gcpsm

go 1.24.0

toolchain go1.25.3

require (
	github.com/zoobzio/fig v0.0.0
	golang.org/x/oauth2 v0.30.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/zoobzio/sentinel v1.0.2 // indirect
)

replace github.com/zoobzio/fig => ../
