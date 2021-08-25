module github.com/startupstation/bantu-core

go 1.16

require (
	github.com/go-gormigrate/gormigrate/v2 v2.0.0
	github.com/soffa-io/soffa-core-go v0.1.3
	github.com/stretchr/testify v1.5.1
	gorm.io/gorm v1.21.13
)

//replace github.com/soffa-io/soffa-core-go => ../soffa-core
