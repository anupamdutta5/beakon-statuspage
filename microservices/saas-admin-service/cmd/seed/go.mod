module github.com/anupamdutta5/statuspage-saas-admin-service/cmd/seed

go 1.21

require (
	github.com/google/uuid v1.3.0
	github.com/lib/pq v1.10.9
	github.com/anupamdutta5/statuspage-saas-admin-service/internal/db/seed v0.0.0
)

replace github.com/anupamdutta5/statuspage-saas-admin-service => ../../..
