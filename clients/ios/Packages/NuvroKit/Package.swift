// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "NuvroKit",
    platforms: [
        .iOS(.v18),
        .macOS(.v15)
    ],
    products: [
        .library(name: "NuvroModels", targets: ["NuvroModels"]),
        .library(name: "NuvroDesignSystem", targets: ["NuvroDesignSystem"]),
        .library(name: "NuvroLocalization", targets: ["NuvroLocalization"]),
        .library(name: "NuvroOnboarding", targets: ["NuvroOnboarding"]),
        .library(name: "NuvroDiscover", targets: ["NuvroDiscover"]),
        .library(name: "NuvroDestination", targets: ["NuvroDestination"]),
        .library(name: "NuvroPlan", targets: ["NuvroPlan"]),
        .library(name: "NuvroTravel", targets: ["NuvroTravel"]),
        .library(name: "NuvroSaved", targets: ["NuvroSaved"]),
        .library(name: "NuvroProfile", targets: ["NuvroProfile"]),
        .library(name: "NuvroAuthentication", targets: ["NuvroAuthentication"]),
        .library(name: "NuvroPermissions", targets: ["NuvroPermissions"]),
        .library(name: "NuvroConnectivity", targets: ["NuvroConnectivity"]),
        .library(name: "NuvroTravelSetup", targets: ["NuvroTravelSetup"]),
        .library(name: "NuvroAPI", targets: ["NuvroAPI"])
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-openapi-generator", from: "1.12.2"),
        .package(url: "https://github.com/apple/swift-openapi-runtime", from: "1.8.2"),
        .package(url: "https://github.com/apple/swift-openapi-urlsession", from: "1.1.0")
    ],
    targets: [
        .target(
            name: "NuvroAPI",
            dependencies: [
                "NuvroModels",
                .product(name: "OpenAPIRuntime", package: "swift-openapi-runtime"),
                .product(name: "OpenAPIURLSession", package: "swift-openapi-urlsession")
            ],
            plugins: [.plugin(name: "OpenAPIGenerator", package: "swift-openapi-generator")]
        ),
        .testTarget(name: "NuvroAPITests", dependencies: ["NuvroAPI"]),
        .target(name: "NuvroModels"),
        .target(name: "NuvroDesignSystem", dependencies: ["NuvroModels"]),
        .target(name: "NuvroLocalization"),
        .target(
            name: "NuvroOnboarding",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization"]
        ),
        .target(
            name: "NuvroDiscover",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels"]
        ),
        .target(
            name: "NuvroDestination",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels"]
        ),
        .target(
            name: "NuvroPlan",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels"]
        ),
        .target(
            name: "NuvroTravel",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels", "NuvroPermissions"]
        ),
        .target(
            name: "NuvroAuthentication",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels", "NuvroAPI"]
        ),
        .testTarget(name: "NuvroModelsTests", dependencies: ["NuvroModels"]),
        .testTarget(name: "NuvroAuthenticationTests", dependencies: ["NuvroAuthentication"]),
        .target(name: "NuvroPermissions", dependencies: ["NuvroDesignSystem"]),
        .target(name: "NuvroConnectivity", dependencies: ["NuvroModels"]),
        .testTarget(name: "NuvroPermissionsTests", dependencies: ["NuvroPermissions"]),
        .target(name: "NuvroTravelSetup", dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels"]),
        .testTarget(name: "NuvroTravelSetupTests", dependencies: ["NuvroTravelSetup"]),
        .testTarget(name: "NuvroConnectivityTests", dependencies: ["NuvroConnectivity"]),
        .target(
            name: "NuvroSaved",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels"]
        ),
        .target(
            name: "NuvroProfile",
            dependencies: ["NuvroDesignSystem", "NuvroLocalization", "NuvroModels", "NuvroPermissions", "NuvroTravelSetup", "NuvroAPI"],
            resources: [.process("Resources")]
        ),
        .testTarget(name: "NuvroTravelTests", dependencies: ["NuvroTravel"]),
        .testTarget(name: "NuvroSavedTests", dependencies: ["NuvroSaved"]),
        .testTarget(name: "NuvroProfileTests", dependencies: ["NuvroProfile"]),
        .testTarget(name: "NuvroDiscoverTests", dependencies: ["NuvroDiscover"]),
        .testTarget(name: "NuvroDestinationTests", dependencies: ["NuvroDestination"]),
        .testTarget(name: "NuvroPlanTests", dependencies: ["NuvroPlan"]),
        .testTarget(name: "NuvroDesignSystemTests", dependencies: ["NuvroDesignSystem"]),
        .testTarget(name: "NuvroLocalizationTests", dependencies: ["NuvroLocalization"]),
        .testTarget(name: "NuvroOnboardingTests", dependencies: ["NuvroOnboarding"])
    ]
)
