module github.com/zoobz-io/fig/gcpsm

go 1.24.0

toolchain go1.25.3

require (
	github.com/zoobz-io/fig v0.0.0
	golang.org/x/oauth2 v0.30.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/zoobz-io/sentinel v1.0.4 // indirect
)

replace github.com/zoobz-io/fig => ../
