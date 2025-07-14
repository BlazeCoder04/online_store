package domain

//go:generate sh -c "rm -rf mocks && mkdir -p mocks"
//go:generate mockgen -source=token.go -destination=mocks/token_adapter_mock.go -package=mocks
