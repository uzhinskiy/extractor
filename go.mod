module github.com/uzhinskiy/extractor

go 1.14

replace github.com/uzhinskiy/extractor/modules/front => ./modules/front

require (
	github.com/uzhinskiy/extractor/modules/front v0.0.0
	gopkg.in/yaml.v2 v2.2.8
)
