module github.com/slonopotamus/annoyer

go 1.27.1

require (
	github.com/anthropics/anthropic-sdk-go v1.84.0
	github.com/ctreminiom/go-atlassian/v2 v2.12.0
	github.com/spf13/cobra v1.10.2
	gitlab.com/gitlab-org/api/client-go/v3 v3.39.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/jsonschema v0.14.0 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/standard-webhooks/standard-webhooks/libraries v0.0.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/time v0.16.0 // indirect
)

// Temporary workaround for https://github.com/anthropics/anthropic-sdk-go/issues/481
// See https://github.com/anthropics/anthropic-sdk-go/pull/482
replace github.com/anthropics/anthropic-sdk-go => github.com/tingly-dev/anthropic-sdk-go v1.84.1-0.20261021030836-be88e657ec1f
