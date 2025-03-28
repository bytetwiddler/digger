module github.com/bytetwiddler/digger

go 1.23.6

require (
	github.com/Graylog2/go-gelf v0.0.0-20170811154226-7ebf4f536d8f
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	go.etcd.io/bbolt v1.4.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bytetwiddler/digger/pkg/config => ./pkg/config

replace github.com/bytetwiddler/digger/pkg/logging => ./pkg/logging
