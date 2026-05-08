module github.com/ilnur/gym-membership-app/telemetry-service

go 1.24

require google.golang.org/grpc v1.70.0

require (
	github.com/ilnur/gym-membership-app/proto v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

replace github.com/ilnur/gym-membership-app/proto => ../proto
