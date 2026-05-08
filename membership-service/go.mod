module github.com/ilnur/gym-membership-app/membership-service

go 1.25.0

require github.com/google/uuid v1.6.0

require (
	github.com/ilnur/gym-membership-app/proto v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/grpc v1.81.0
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/ilnur/gym-membership-app/proto => ../proto
