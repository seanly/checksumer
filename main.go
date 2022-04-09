package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// define the input API schema as a struct

type ChecksumSpec struct {
	Key 	string `yaml:"key"`
	Path 	string `yaml:"path"`
	Target  *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
}

type SelectorSpec struct {
	Target      *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
	FieldSpec 	types.FieldSpec `json:"fieldSpec,omitempty" yaml:"fieldSpec,omitempty"`
}

type Checksumer struct {
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`

	Spec struct {
		Checksum []ChecksumSpec `json:"checksum,omitempty" yaml:"checksum"`
		Selectors []SelectorSpec `json:"selectors,omitempty" yaml:"selectors,omitempty"`
	} `yaml:"spec"`
}

type Filter struct {
	checksumMap map[string]string
	FieldSpec   types.FieldSpec `json:"fieldSpec,omitempty" yaml:"fieldSpec,omitempty"`
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(f.run)).Filter(nodes)
}

func (f Filter) run(node *yaml.RNode) (*yaml.RNode, error) {

	var trackableSetter filtersutil.TrackableSetter
	keys := yaml.SortedMapKeys(f.checksumMap)
	for _, k := range keys {
		if err := node.PipeE(fieldspec.Filter{
			FieldSpec: f.FieldSpec,
			SetValue:  trackableSetter.SetEntry(k, f.checksumMap[k], yaml.NodeTagMap),
			CreateKind: yaml.MappingNode,
			CreateTag: yaml.NodeTagMap,
		}); err != nil {
			return nil, err
		}
	}
	return node, nil
}

func main() {

	config := new(Checksumer)

	fn := func(items []*yaml.RNode)([]*yaml.RNode, error) {

		p := provider.NewDefaultDepProvider()
		resmapFactory := resmap.NewFactory(p.GetResourceFactory())
		resMap, err := resmapFactory.NewResMapFromRNodeSlice(items)

		computedKVs := make(map[string]string, 0)
		for _, checksumSpec := range config.Spec.Checksum {
			resources, err := resMap.Select(*checksumSpec.Target)
			if err != nil {
				return items, err
			}

			for _, res := range resources {
				if res.IsNilOrEmpty() {
					continue
				}
				content, err := res.AsYAML()
				if err != nil {
					return items, err
				}
				computedKVs[checksumSpec.Key] = SHA1Sum(content)
			}
		}

		for _, selector := range config.Spec.Selectors {
			resources, err := resMap.Select(*selector.Target)
			if err != nil {
				return items, err
			}

			for _, res := range resources {
				err = res.ApplyFilter(Filter{
					checksumMap: computedKVs,
					FieldSpec: selector.FieldSpec,
				})
			}
		}
		return items, err
	}

	p := framework.SimpleProcessor{Config: config,Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func SHA1Sum(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}