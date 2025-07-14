package domain

//go:generate sh -c "rm -rf mocks && mkdir -p mocks"
//go:generate mockgen -source=user.go -destination=mocks/user_repository_mock.go -package=mocks
