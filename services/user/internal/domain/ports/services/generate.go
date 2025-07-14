package domain

//go:generate sh -c "rm -rf mocks && mkdir -p mocks"
//go:generate mockgen -source=auth.go -destination=mocks/auth_service_mock.go -package=mocks
