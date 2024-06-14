package proto

import (
	"context"
	"fmt"
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
	"path/filepath"
	"strings"
)

type descriptorsManager struct {
	descriptors map[protoreflect.FullName]protoreflect.MethodDescriptor
}

func NewDescriptorsManager(cfg *config.Global, testCases config.TestCases) (DescriptorsManager, error) {
	manager := &descriptorsManager{
		descriptors: make(map[protoreflect.FullName]protoreflect.MethodDescriptor),
	}

	files, err := collectFiles(cfg.ProtoSources, cfg.ProtoRoot)
	c := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: append(cfg.ProtoImports, cfg.ProtoRoot),
		}),
	}
	if err != nil {
		return nil, errors.Wrap(err, "error collecting proto sources")
	}

	compiled, err := c.Compile(context.Background(), files...)
	if err != nil {
		return nil, errors.Wrap(err, "error compiling proto sources")
	}

	err = updateDescriptors(manager, testCases, compiled)
	if err != nil {
		return nil, errors.Wrap(err, "error updating descriptors")
	}

	return manager, nil
}

func collectFiles(sources []string, root string) ([]string, error) {
	files := make([]string, 0)
	var err error

	if len(sources) == 0 {
		sources, err = fileList(root)
		if err != nil {
			return nil, errors.Wrap(err, "error reading proto sources")
		}
	}

	for _, source := range sources {
		info, err := os.Stat(root + "/" + source)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path %s does not exist", source)
		}
		if err != nil {
			return nil, errors.Wrap(err, "error reading proto source")
		}

		if info.IsDir() {
			protoFiles, err := filepath.Glob(root + "/" + source + "/*.proto")
			if err != nil {
				return nil, errors.Wrap(err, "error reading proto source")
			}
			for _, file := range protoFiles {
				files = append(files, strings.Replace(file, root+"/", "", 1))
			}
		} else {
			files = append(files, source)
		}
	}

	return files, nil
}

func fileList(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, strings.Replace(path, root+"/", "", 1))
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "error reading proto files")
	}

	return files, nil
}

func updateDescriptors(manager *descriptorsManager, testCases config.TestCases, compiled linker.Files) error {
	for _, testCase := range testCases {
		for _, step := range testCase.Steps {
			fullName := step.Service.Service + "." + step.Method
			fullNameReflect := protoreflect.FullName(fullName)
			if _, ok := manager.descriptors[fullNameReflect]; ok {
				continue
			}

			var d protoreflect.Descriptor
			for _, fd := range compiled {
				if !fullNameReflect.IsValid() {
					return fmt.Errorf("method %s is not valid", fullName)
				}

				if d = fd.FindDescriptorByName(fullNameReflect); d != nil {
					break
				}
			}
			if d == nil {
				return fmt.Errorf("method %s not found in sources", fullName)
			}

			manager.descriptors[fullNameReflect] = d.(protoreflect.MethodDescriptor)
		}
	}

	return nil
}

func (d *descriptorsManager) GetDescriptor(name protoreflect.FullName) protoreflect.MethodDescriptor {
	return d.descriptors[name]
}
