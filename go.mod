module github.com/uzhinskiy/extractor

go 1.14

replace github.com/uzhinskiy/extractor/modules/front => ./modules/front

replace github.com/uzhinskiy/extractor/modules/router => ./modules/router

require (
	github.com/uzhinskiy/extractor/modules/front v0.0.0 // indirect
	github.com/uzhinskiy/extractor/modules/router v0.0.0
	github.com/uzhinskiy/lib.go v0.0.0-20201103093805-d91c0db6a10b // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)
