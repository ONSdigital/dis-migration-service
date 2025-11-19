# dis-migration-service

A Go API for data migration.

## Getting started

* Run `make debug` to run application on <http://localhost:30100>
* Run `make debug-watch` to have your changes [rebuild the application](#watch-for-changes) that is running
* Run `make help` to see full list of make targets

## Tools

To run some of our tests you will need additional tooling:

### Watch for changes

We use `reflex` to do rebuilds, which you will [need to install](https://github.com/cespare/reflex).

### Audit

We use `dis-vulncheck` to do auditing, which you will [need to install](https://github.com/ONSdigital/dis-vulncheck).

#### Linting

We use v2 of golangci-lint, which you will [need to install](https://golangci-lint.run/docs/welcome/install).

## Dependencies

* No further dependencies other than those defined in `go.mod`

## Validating Specification

To validate the swagger specification you can do this via:

```sh
make validate-specification
```

To run this, you will need to run Node > v22 and have [redocly CLI v2](https://github.com/Redocly/redocly-cli) installed:

```sh
npm install -g @redocly/cli
```

## Configuration

| Environment variable                      | Default                | Description                                                                                                        |
|-------------------------------------------|------------------------|--------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                                 | :30100                 | The host and port to bind to                                                                                       |
| DATASET_API_URL                           | localhost:20000        | Address for Dataset API                                                                                            |
| DEFAULT_LIMIT                             | 10                     | Default limit parameter for paginated endpoints                                                                    |
| DEFAULT_MAX_LIMIT                         | 100                    | Default max limit for paginated endpoints                                                                          |
| DEFAULT_OFFSET                            | 0                      | Default offset parameter for paginated endpoints                                                                   |
| ENABLE_MOCK_CLIENTS                       | false                  | Boolean to inject mock clients to allow for faster development                                                     |
| FILES_API_URL                             | localhost:26900        | Address for File API                                                                                               |
| GRACEFUL_SHUTDOWN_TIMEOUT                 | 5s                     | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL                      | 30s                    | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT              | 90s                    | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| MIGRATION_SERVICE_URL                     | http://localhost:30100 | Host address used for deriving HATEOS link defaults                                                                |
| OTEL_EXPORTER_OTLP_ENDPOINT               | localhost:4317         | Endpoint for OpenTelemetry service                                                                                 |
| OTEL_SERVICE_NAME                         | dis-migration-service  | Label of service for OpenTelemetry service                                                                         |
| OTEL_BATCH_TIMEOUT                        | 5s                     | Timeout for OpenTelemetry                                                                                          |
| OTEL_ENABLED                              | false                  | Feature flag to enable OpenTelemetry                                                                               |
| REDIRECT_API_URL                          | localhost:29900        | Address for the Redirect API                                                                                       |
| UPLOAD_SERVICE_URL                        | localhost:25100        | Address for Upload Service                                                                                         |
| ZEBEDEE_URL                               | localhost:8082         | Address for Zebedee                                                                                                |
| AUTHORISATION_ENABLED                     | false                  | Feature flag to enable authorisation to be required on endpoints                                                   |
| JWT_VERIFICATION_PUBLIC_KEYS              |                        | A map of public key names and values                                                                               |
| PERMISSIONS_API_URL                       | localhost:25400        | Endpoint for the Permissions API                                                                                   |
| PERMISSIONS_CACHE_UPDATE_INTERVAL         | 1 minute               | The set length of time until the cached permissions will be refreshed from the origin server                       |
| PERMISSIONS_MAX_CACHE_TIME                | 5 minutes              | The maximum length of time that permissions can be cached before they must be refreshed from the origin            |
| IDENTITY_WEB_KEY_SET_URL                  | localhost:25600        | Endpoint for the Identity API                                                                                      |
| AUTHORISATION_IDENTITY_CLIENT_MAX_RETRIES | 2                      | The maximum number of times that the service tries to connect to the Identity API                                  |

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
