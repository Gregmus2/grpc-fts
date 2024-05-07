package config_test

import (
	"flag"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestNewServices_NoFile(t *testing.T) {
	flagSet := flag.NewFlagSet("", 0)
	flagSet.String("configs", ".", "path to configs directory")
	err := flagSet.Set("configs", "testdata")
	if err != nil {
		t.Fatal(err)
	}

	ctx := cli.NewContext(nil, flagSet, nil)
	_, err = config.NewServices(ctx)

	assert.ErrorAs(t, err, &models.UserErr{})
}

func TestNewServices_InvalidFile(t *testing.T) {
	flagSet := flag.NewFlagSet("", 0)
	flagSet.String("configs", ".", "path to configs directory")

	file, err := os.Create("services.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("services.yaml")
	_, err = file.WriteString("invalid yaml")
	if err != nil {
		t.Fatal(err)
	}

	ctx := cli.NewContext(nil, flagSet, nil)
	_, err = config.NewServices(ctx)

	assert.ErrorContains(t, err, "error parsing service config")
}

func TestNewServices_Positive(t *testing.T) {
	flagSet := flag.NewFlagSet("", 0)
	flagSet.String("configs", ".", "path to configs directory")

	file, err := os.Create("services.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("services.yaml")
	_, err = file.WriteString(`
foo:
  service: public.FooService
  address: "foo:9000"
  metadata:
    bar: 123
`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := cli.NewContext(nil, flagSet, nil)
	services, err := config.NewServices(ctx)

	assert.NoError(t, err)
	assert.Contains(t, services, "foo")
	assert.Equal(t, services["foo"].Service, "public.FooService")
	assert.Equal(t, services["foo"].Address, "foo:9000")
	assert.Contains(t, services["foo"].Metadata, "bar")
	assert.Equal(t, services["foo"].Metadata["bar"], "123")
}
