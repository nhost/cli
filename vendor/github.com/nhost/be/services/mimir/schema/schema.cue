package config

import (
	"net"
	"strings"
)

#Config: {
	global:    #Global
	hasura:    #Hasura
	functions: #Functions
	auth:      #Auth
	provider:  #Provider
	storage:   #Storage
}

#Global: {
	// User-defined environment variables that are spread over all services
	environment: [...#EnvironmentVariable] | *[]
	// Name of the application
	name: string | *"Nhost application"
}

#EnvironmentVariable: {
	name:  =~"[a-zA-Z_]{1,}[a-zA-Z0-9_]*"
	value: string
}

#Hasura: {
	version: "v2.10.1" | *"v2.15.2"

	jwtSecrets: [#JWTSecret]
	adminSecret:   string
	webhookSecret: string

	settings: {
		enableRemoteSchemaPermissions: bool | *false
	}
}

#Storage: {
	version: "0.2.3" | "0.3.0" | *"0.3.1"
}

#Functions: {
	node: {
		version: 16
	}
}

#Auth: {
	version: =~"sha-.*" | "0.13.2" | "0.16.2" | *"0.17.0"

	redirections: {
		clientUrl: #Url | *"http://localhost:3000"
		// We should implement wildcards soon, so the #Url type should not be used here
		allowedUrls: [...string]
	}

	signUp: {
		enabled: bool | *true
	}

	user: {
		roles: {
			default: #UserRole | *"user"
			allowed: [ ...#UserRole] | *[default, "me"]
		}
		locale: {
			default: #Locale | *"en"
			allowed: [...#Locale] | *[default]
		}

		gravatar: {
			enabled: bool | *true
			default: "404" | "mp" | "identicon" | "monsterid" | "wavatar" | "retro" | "robohash" | *"blank"
			rating:  "pg" | "r" | "x" | *"g"
		}
		email: {
			allowed: [...#Email]
			blocked: [...#Email]

		}
		emailDomains: {
			allowed: [...string & net.FQDN]
			blocked: [...string & net.FQDN]
		}
	}

	session: {
		accessToken: {
			expiresIn:    uint32 | *900
			customClaims: [...{
				key:   =~"[a-zA-Z_]{1,}[a-zA-Z0-9_]*"
				value: string
			}] | *[]
		}

		refreshToken: {
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
			// Will be implemented soon
			// defaults: {
			//  signUp: {
			//   enabled: true
			//  }
			// }
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
			}
			azuread: {
				#StandardOauthProvider
				tenant: string | *"common"
			}
			bitbucket: #StandardOauthProvider
			discord:   #StandardOauthProvider
			facebook:  #StandardOauthProvider
			github:    #StandardOauthProvider
			gitlab:    #StandardOauthProvider
			google:    #StandardOauthProvider
			linkedin:  #StandardOauthProvider
			spotify:   #StandardOauthProvider
			strava:    #StandardOauthProvider
			twitch:    #StandardOauthProvider
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
			windowslive: #StandardOauthProvider
			workos: {
				#StandardOauthProvider
				connection?:       string
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
	scope?: [...string]
}

#Provider: {
	smtp: #Smtp
	sms?: #Sms
}

#Smtp: {
	user:     string
	password: string
	sender:   string | *""
	host:     string & net.FQDN | net.IP
	port:     #Port | *1025
	secure:   bool | *false
	method:   "LOGIN" | "GSSAPI" | "GSSAPI" | "DIGEST-MD5" | "MD5" | "CRAM-MD5" | "OAUTH10A" | "OAUTHBEARER" | "XOAUTH2" | *"PLAIN"
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

// * https://hasura.io/docs/latest/auth/authentication/jwt/
#JWTSecret:
	({
		type: "HS384" | "HS512" | "RS256" | "RS384" | "RS512" | "Ed25519" | *"HS256"
		key:  string
	} |
	{
		jwk_url: #Url | *null
	}) &
	{
		// TODO what's this?
		// claims_map?: {}
		claims_format?: "stringified_json" | *"json"
		audience?:      string
		issuer?:        string
		allowed_skew?:  uint32
		header?:        string
	} &
	({
		claims_namespace: string | *"https://hasura.io/jwt/claims"
	} |
	{
		claims_namespace_path: string
	} | *{})

#SystemConfig: {
	auth: {
		email: {
			templates: {
				s3Key?: string
			}
		}
	}

	postgres: {
		version:            string
		connectionSettings: #PostgresConnectionSettings
	}
}

#PostgresConnectionSettings: ({
	type:     "rds"
	database: string
	host:     string
	port:     #Port
	user:     string
	password: string
} | {
	type:     "postgres-in-k8s"
	database: string
	host:     string
	password: string
})
