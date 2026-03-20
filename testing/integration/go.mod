module github.com/zoobz-io/fig/testing/integration

go 1.24.0

toolchain go1.25.3

require (
	github.com/aws/aws-sdk-go-v2/config v1.32.7
	github.com/aws/aws-sdk-go-v2/credentials v1.19.7
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.41.1
	github.com/zoobz-io/fig v0.0.0
	github.com/zoobz-io/fig/awssm v0.0.0
	github.com/zoobz-io/fig/gcpsm v0.0.0
	github.com/zoobz-io/fig/vault v0.0.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.41.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.6 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/zoobz-io/sentinel v1.0.4 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
)

replace (
	github.com/zoobz-io/fig => ../../
	github.com/zoobz-io/fig/awssm => ../../awssm
	github.com/zoobz-io/fig/gcpsm => ../../gcpsm
	github.com/zoobz-io/fig/vault => ../../vault
)
