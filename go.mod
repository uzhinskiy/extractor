module github.com/uzhinskiy/extractor

replace (
	github.com/uzhinskiy/extractor/modules/front => ./modules/front
)

require (
	github.com/uzhinskiy/extractor/modules/config v0.0.0
	gopkg.in/yaml.v2 v2.2.8
)
