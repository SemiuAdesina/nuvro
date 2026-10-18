# Mobile contract generation

```
contracts/openapi/v1/openapi.yaml
  ├── oapi-codegen strict Chi Go server
  ├── Apple Swift OpenAPI Generator + URLSession transport
  └── type-safe Kotlin generated client
```

Generate only released operations. The Swift client is generated at build time by the
`OpenAPIGenerator` package plugin inside the `NuvroAPI` target of `NuvroKit`, from the
bundled document at `clients/ios/Packages/NuvroKit/Sources/NuvroAPI/openapi.json`. That file
is produced by `scripts/ios-contract-sync.sh` from the released contract and checked by the
`ios-contract` gate in `scripts/verify.sh`, so the client can never drift from the server.
Nothing generated is committed; the plugin output lives in `.build`. The Android Gradle
project is not initialised, so no Kotlin client is generated yet.

Pinned: swift-openapi-generator 1.12, swift-openapi-runtime 1.8, swift-openapi-urlsession 1.1.
Generated types stay inside `NuvroAPI`; every other module sees plain value types through the
`SessionGateway` and `AccountGateway` protocols, which is where generated models are mapped.
When the Android client is ready, select a maintained Kotlin generator transport that can be
built by the Android toolchain.
Validate OpenAPI 3.1 nullable PATCH, enums, int64 money, RFC3339 and Problem Details
with client fixtures before promotion. Generated files live under explicit Generated
locations and are never hand-edited. Map to client presentation models where needed.

The Phase 0 prototype schemas and manually modeled Swift fixtures are not a released
network contract. In particular, the backend now uses camelCase and RFC 9457;
prototype snake_case payloads must not be consumed as production responses.
Do not duplicate hundreds of networking DTOs by hand during the integration.
