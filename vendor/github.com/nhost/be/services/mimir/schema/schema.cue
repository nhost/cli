package schema

import (
	"list"
	"math"
	"net"
	"strings"
)

// main entrypoint to the configuration
#Config: {
	// Global configuration that applies to all services
	global: #Global

	// Configuration for hasura
	hasura: #Hasura

	// Configuration for functions service
	functions: #Functions

	// Configuration for auth service
	auth: #Auth

	// Configuration for postgres service
	postgres: #Postgres

	// Configuration for third party providers like SMTP, SMS, etc.
	provider: #Provider

	// Configuration for storage service
	storage: #Storage

	// Configuration for observability service
	observability: #Observability

	_totalResourcesCPU: (
				hasura.resources.replicas*hasura.resources.compute.cpu +
		auth.resources.replicas*auth.resources.compute.cpu +
		storage.resources.replicas*storage.resources.compute.cpu +
		postgres.resources.replicas*postgres.resources.compute.cpu) @cuegraph(skip)

	_totalResourcesMemory: (
				hasura.resources.replicas*hasura.resources.compute.memory +
		auth.resources.replicas*auth.resources.compute.memory +
		storage.resources.replicas*storage.resources.compute.memory +
		postgres.resources.replicas*postgres.resources.compute.memory) @cuegraph(skip)

	_validateResourcesTotalCpuMemoryRatioMustBe1For2: (
								_totalResourcesCPU*2.048 & _totalResourcesMemory*1.0) @cuegraph(skip)

	_validateResourcesTotalCpuMin1000: (
						hasura.resources.compute.cpu+
		auth.resources.compute.cpu+
		storage.resources.compute.cpu+
		postgres.resources.compute.cpu) >= 1000 & true @cuegraph(skip)

	_validateAllResourcesAreSetOrNot: (
						(hasura.resources == _|_) ==
		(auth.resources == _|_) ==
		(storage.resources == _|_) ==
		(postgres.resources == _|_) ) & true @cuegraph(skip)
}

// Global configuration that applies to all services
#Global: {
	// User-defined environment variables that are spread over all services
	environment: [...#EnvironmentVariable] | *[]
}

#EnvironmentVariable: {
	// Name of the environment variable
	name: =~"(?i)^[a-z_]{1,}[a-z0-9_]*" & !~"(?i)^NHOST_" & !~"(?i)^HASURA_"
	// Value of the environment variable
	value: string
}

// Resource configuration for a service
#Resources: {
	compute: {
		// milicpus, 1000 milicpus = 1 cpu
		cpu: uint32 & >=250 & <=15000
		// MiB: 128MiB to 30GiB
		memory: uint32 & >=128 & <=30720

		// validate CPU steps of 250 milicpus
		_validateCPUSteps250: (mod(cpu, 250) == 0) & true @cuegraph(skip)

		// validate memory steps of 128 MiB
		_validateMemorySteps128: (mod(memory, 128) == 0) & true @cuegraph(skip)
	}

	// Number of replicas for a service
	replicas: uint8 & >=1 & <=10

	_validateMultipleReplicasRatioMustBe1For2: (
							replicas == 1 |
		(compute.cpu*2.048 == compute.memory)) & true @cuegraph(skip)
}

// Configuration for hasura service
#Hasura: {
	// Version of hasura, you can see available versions in the URL below:
	// https://hub.docker.com/r/hasura/graphql-engine/tags
	version: string | *"v2.33.4-ce"

	// JWT Secrets configuration
	jwtSecrets: [#JWTSecret]

	// Admin secret
	adminSecret: string

	// Webhook secret
	webhookSecret: string

	// Configuration for hasura services
	// Reference: https://hasura.io/docs/latest/deployment/graphql-engine-flags/reference/
	settings: {
		// HASURA_GRAPHQL_CORS_DOMAIN
		corsDomain: [...#Url] | *["*"]
		// HASURA_GRAPHQL_DEV_MODE
		devMode: bool | *true
		// HASURA_GRAPHQL_ENABLE_ALLOWLIST
		enableAllowList: bool | *false
		// HASURA_GRAPHQL_ENABLE_CONSOLE
		enableConsole: bool | *true
		// HASURA_GRAPHQL_ENABLE_REMOTE_SCHEMA_PERMISSIONS
		enableRemoteSchemaPermissions: bool | *false
		// HASURA_GRAPHQL_ENABLED_APIS
		enabledAPIs: [...#HasuraAPIs] | *["metadata", "graphql", "pgdump", "config"]

		// HASURA_GRAPHQL_LIVE_QUERIES_MULTIPLEXED_REFETCH_INTERVAL
		liveQueriesMultiplexedRefetchInterval: uint32 | *1000
	}

	logs: {
		// HASURA_GRAPHQL_LOG_LEVEL
		level: "debug" | "info" | "error" | *"warn"
	}

	events: {
		// HASURA_GRAPHQL_EVENTS_HTTP_POOL_SIZE
		httpPoolSize: uint32 & >=1 & <=100 | *100
	}

	// Resources for the service
	resources?: #Resources
}

// APIs for hasura
#HasuraAPIs: "metadata" | "graphql" | "pgdump" | "config"

// Configuration for storage service
#Storage: {
	// Version of storage service, you can see available versions in the URL below:
	// https://hub.docker.com/r/nhost/hasura-storage/tags
	//
	// Releases:
	//
	// https://github.com/nhost/hasura-storage/releases
	version: string | *"0.4.0"

	// Resources for the service
	resources?: #Resources

	antivirus?: {
		server: "tcp://run-clamav:3310"
	}
}

// Configuration for functions service
#Functions: {
	node: {
		version: 18
	}
}

// Configuration for postgres service
#Postgres: {
	// Version of postgres, you can see available versions in the URL below:
	// https://hub.docker.com/r/nhost/postgres/tags
	version: string | *"14.6-20230927-1"

	// Resources for the service
	resources?: #Resources & {
		replicas: 1
	}

	settings?: {
		jit:                           "off" | "on" | *"on"
		maxConnections:                int32 | *100
		sharedBuffers:                 string | *"128MB"
		effectiveCacheSize:            string | *"4GB"
		maintenanceWorkMem:            string | *"64MB"
		checkpointCompletionTarget:    number | *0.9
		walBuffers:                    string | *"-1"
		defaultStatisticsTarget:       int32 | *100
		randomPageCost:                number | *4.0
		effectiveIOConcurrency:        int32 | *1
		workMem:                       string | *"4MB"
		hugePages:                     string | *"try"
		minWalSize:                    string | *"80MB"
		maxWalSize:                    string | *"1GB"
		maxWorkerProcesses:            int32 | *8
		maxParallelWorkersPerGather:   int32 | *2
		maxParallelWorkers:            int32 | *8
		maxParallelMaintenanceWorkers: int32 | *2
	}
}

// Configuration for auth service
// You can find more information about the configuration here:
// https://github.com/nhost/hasura-auth/blob/main/docs/environment-variables.md
#Auth: {
	// Version of auth, you can see available versions in the URL below:
	// https://hub.docker.com/r/nhost/hasura-auth/tags
	//
	// Releases:
	//
	// https://github.com/nhost/hasura-auth/releases
	version: string | *"0.21.2"

	// Resources for the service
	resources?: #Resources

	redirections: {
		// AUTH_CLIENT_URL
		clientUrl: #Url | *"http://localhost:3000"
		// AUTH_ACCESS_CONTROL_ALLOWED_REDIRECT_URLS
		allowedUrls: [...string]
	}

	signUp: {
		// Inverse of AUTH_DISABLE_NEW_USERS
		enabled: bool | *true
	}

	user: {
		roles: {
			// AUTH_USER_DEFAULT_ROLE
			default: #UserRole | *"user"
			// AUTH_USER_DEFAULT_ALLOWED_ROLES
			allowed: [ ...#UserRole] | *[default, "me"]
		}
		locale: {
			// AUTH_LOCALE_DEFAULT
			default: #Locale | *"en"
			// AUTH_LOCALE_ALLOWED_LOCALES
			allowed: [...#Locale] | *[default]
		}

		gravatar: {
			// AUTH_GRAVATAR_ENABLED
			enabled: bool | *true
			// AUTH_GRAVATAR_DEFAULT
			default: "404" | "mp" | "identicon" | "monsterid" | "wavatar" | "retro" | "robohash" | *"blank"
			// AUTH_GRAVATAR_RATING
			rating: "pg" | "r" | "x" | *"g"
		}
		email: {
			// AUTH_ACCESS_CONTROL_ALLOWED_EMAILS
			allowed: [...#Email]
			// AUTH_ACCESS_CONTROL_BLOCKED_EMAILS
			blocked: [...#Email]

		}
		emailDomains: {
			// AUTH_ACCESS_CONTROL_ALLOWED_EMAIL_DOMAINS
			allowed: [...string & net.FQDN]
			// AUTH_ACCESS_CONTROL_BLOCKED_EMAIL_DOMAINS
			blocked: [...string & net.FQDN]
		}
	}

	session: {
		accessToken: {
			// AUTH_ACCESS_TOKEN_EXPIRES_IN
			expiresIn: uint32 | *900
			// AUTH_JWT_CUSTOM_CLAIMS
			customClaims: [...{
				key:   =~"[a-zA-Z_]{1,}[a-zA-Z0-9_]*"
				value: string
			}] | *[]
		}

		refreshToken: {
			// AUTH_REFRESH_TOKEN_EXPIRES_IN
			expiresIn: uint32 | *43200
		}

	}

	method: {
		anonymous: {
			enabled: bool | *false
		}

		emailPasswordless: {
			enabled: bool | *false
		}

		emailPassword: {
			// Disabling email+password sign in is not implmented yet
			// enabled: bool | *true
			hibpEnabled:               bool | *false
			emailVerificationRequired: bool | *true
			passwordMinLength:         uint8 & >=3 | *9
		}

		smsPasswordless: {
			enabled: bool | *false
		}

		oauth: {
			apple: {
				enabled: bool | *false
				if enabled {
					clientId:   string
					keyId:      string
					teamId:     string
					privateKey: string
				}
				if !enabled {
					clientId?:   string
					keyId?:      string
					teamId?:     string
					privateKey?: string
				}
				scope?: [...string]
			}
			azuread: {
				#StandardOauthProvider
				tenant: string | *"common"
			}
			bitbucket: #StandardOauthProvider
			discord:   #StandardOauthProviderWithScope
			facebook:  #StandardOauthProviderWithScope
			github:    #StandardOauthProviderWithScope
			gitlab:    #StandardOauthProviderWithScope
			google:    #StandardOauthProviderWithScope
			linkedin:  #StandardOauthProviderWithScope
			spotify:   #StandardOauthProviderWithScope
			strava:    #StandardOauthProviderWithScope
			twitch:    #StandardOauthProviderWithScope
			twitter: {
				enabled: bool | *false
				if enabled {
					consumerKey:    string
					consumerSecret: string
				}
				if !enabled {
					consumerKey?:    string
					consumerSecret?: string
				}
			}
			windowslive: #StandardOauthProviderWithScope
			workos: {
				#StandardOauthProvider
				connection?:   string
				organization?: string
			}
		}

		webauthn: {
			enabled: bool | *false
			if enabled {
				relyingParty: {
					name:    string
					origins: [...#Url] | *[redirections.clientUrl]
				}
			}
			if !enabled {
				relyingParty?: {
					name?:    string
					origins?: [...#Url] | *[redirections.clientUrl]
				}
			}
			attestation: {
				timeout: uint32 | *60000
			}
		}
	}

	totp: {
		enabled: bool | *false
		if enabled {
			issuer: string
		}
		if !enabled {
			issuer?: string
		}
	}

}

#StandardOauthProvider: {
	enabled: bool | *false
	if enabled {
		clientId:     string
		clientSecret: string
	}
	if !enabled {
		clientId?:     string
		clientSecret?: string
	}
}

#StandardOauthProviderWithScope: {
	enabled: bool | *false
	if enabled {
		clientId:     string
		clientSecret: string
	}
	if !enabled {
		clientId?:     string
		clientSecret?: string
	}
	scope?: [...string]
}

#Provider: {
	smtp?: #Smtp
	sms?:  #Sms
}

#Smtp: {
	user:     string
	password: string
	sender:   string
	host:     string & net.FQDN | net.IP
	port:     #Port
	secure:   bool
	method:   "LOGIN" | "GSSAPI" | "GSSAPI" | "DIGEST-MD5" | "MD5" | "CRAM-MD5" | "OAUTH10A" | "OAUTHBEARER" | "XOAUTH2" | "PLAIN"
}

#Sms: {
	provider:           "twilio"
	accountSid:         string
	authToken:          string
	messagingServiceId: string
}

#UserRole: string
#Url:      string
#Port:     uint16
#Email:    =~"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
#Locale:   string & strings.MinRunes(2) & strings.MaxRunes(2)

// See https://hasura.io/docs/latest/auth/authentication/jwt/
#JWTSecret:
	({
		type: "HS384" | "HS512" | "RS256" | "RS384" | "RS512" | "Ed25519" | *"HS256"
		key:  string
	} |
	{
		jwk_url: #Url | *null
	}) &
	{
		claims_format?: "stringified_json" | *"json"
		audience?:      string
		issuer?:        string
		allowed_skew?:  uint32
		header?:        string
	} & {
		claims_map?: [...#ClaimMap]

	} &
	({
		claims_namespace: string | *"https://hasura.io/jwt/claims"
	} |
	{
		claims_namespace_path: string
	} | *{})

#ClaimMap: {
	claim: string
	{
		value: string
	} | {
		path:     string
		default?: string
	}
} & {

}

#SystemConfig: {
	auth: {
		email: {
			templates: {
				s3Key?: string
			}
		}
	}

	postgres: {
		enabled: bool | *true
		if enabled {
			database: string
		}
		if !enabled {
			database?: string
		}
		connectionString: {
			backup:  string
			hasura:  string
			auth:    string
			storage: string
		}
	}
}

#Observability: {
	grafana: #Grafana
}

#Grafana: {
	adminPassword: string
}

#RunServicePort: {
	port:                              #Port
	type:                              "http" | "tcp" | "udp"
	publish:                           bool | *false
	_publish_supported_only_over_http: (
						publish == false || type == "http" ) & true @cuegraph(skip)
}

#RunServiceName: =~"^[a-z]([-a-z0-9]*[a-z0-9])?$" & strings.MinRunes(1) & strings.MaxRunes(12)

// Resource configuration for a service
#RunServiceResources: {
	compute: {
		// milicpus, 1000 milicpus = 1 cpu
		cpu: uint32 & >=62 & <=14000
		// MiB: 128MiB to 30GiB
		memory: uint32 & >=128 & <=28720

		// validate memory steps of 128 MiB
		_validateMemorySteps128: (mod(memory, 128) == 0) & true @cuegraph(skip)
	}

	storage: [...{
		name:     #RunServiceName       // name of the volume, changing it will cause data loss
		capacity: uint32 & >=1 & <=1000 // GiB
		path:     string
	}] | *[]
	_storage_name_must_be_unique: list.UniqueItems([ for s in storage {s.name}]) & true @cuegraph(skip)
	_storage_path_must_be_unique: list.UniqueItems([ for s in storage {s.path}]) & true @cuegraph(skip)

	// Number of replicas for a service
	replicas: uint8 & <=10

	_replcas_cant_be_greater_than_1_when_using_storage: (len(storage) == 0 | (len(storage) > 0 & replicas <= 1)) & true @cuegraph(skip)

	_validate_cpu_memory_ratio_must_be_1_for_2: (math.Abs(compute.memory-compute.cpu*2.048) <= 1.024) & true @cuegraph(skip)
}

#RunServiceImage: {
	image: string
}

#RunServiceConfig: {
	name:  #RunServiceName
	image: #RunServiceImage
	command: [...string]
	environment: [...#EnvironmentVariable] | *[]
	ports?:      [...#RunServicePort] | *[]
	resources:   #RunServiceResources
}
