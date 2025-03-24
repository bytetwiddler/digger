module github.com/bytetwiddler/digger

go 1.23.6

require (
	github.com/Graylog2/go-gelf v0.0.0-20170811154226-7ebf4f536d8f
	github.com/gofrs/flock v0.12.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/sys v0.26.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bytetwiddler/digger/pkg/config => ./pkg/config

replace github.com/bytetwiddler/digger/pkg/logging => ./pkg/logging
