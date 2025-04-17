module glm-xmpp

go 1.23.7

toolchain go1.23.8

require (
	golang.org/x/net v0.37.0
	mellium.im/sasl v0.3.2
	mellium.im/xmlstream v0.15.4
	mellium.im/xmpp v0.22.0
	pain.agency/oasis-sdk v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
	mellium.im/reader v0.1.0 // indirect
)

replace pain.agency/oasis-sdk => ./oasis-sdk
