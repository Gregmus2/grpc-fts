package logic

import (
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"os"
)

type setupHelper struct {
	dir string
}

func NewSetupHelper(ctx config.ContextWrapper) SetupHelper {
	return &setupHelper{
		dir: ctx.DirectoryFlag(),
	}
}

func (s setupHelper) Setup() error {
	if err := os.Mkdir(s.dir+"/test-cases", os.ModePerm); err != nil {
		return errors.Wrap(err, "error creating test-cases directory")
	}
	if _, err := os.Create(s.dir + "/test-cases/.gkeep"); err != nil {
		return errors.Wrap(err, "error creating .gkeep file")
	}

	file, err := os.Create(s.dir + "/global.yaml")
	if err != nil {
		return errors.Wrap(err, "error creating global.yaml")
	}
	_, err = file.WriteString(
		`proto_root: "path to root directory with your proto files"
proto_sources:
  - "it can be relative path (in proto root) to directory with proto files"
  - "or it can be relative path (in proto root) to some specific proto file"
proto_imports:
  - "path to additional proto imports, like google protobuf utilities for example"
`)
	if err != nil {
		return errors.Wrap(err, "error writing to global.yaml")
	}

	file, err = os.Create(s.dir + "/services.yaml")
	if err != nil {
		return errors.Wrap(err, "error creating services.yaml")
	}
	_, err = file.WriteString(`foo:
    # full service name with package
    service: package1.Foo
    # address to your service with port included
    address: "foo:9000"

bar:
    service: package2.Bar
    address: "bar:9000"
    # you can provide any metadata that your service requires
    metadata:
        authorization: $authorization
`)

	file, err = os.Create(s.dir + "/variables.yaml")
	if err != nil {
		return errors.Wrap(err, "error creating variables.yaml")
	}
	_, err = file.WriteString(`# variables can be used in services metadata or in test-cases (requests, responses)
authorization: some-token
`)

	return nil
}
