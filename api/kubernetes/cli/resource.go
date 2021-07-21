package cli

import (
	"bytes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func GenerateYAML(obj runtime.Object) (string, error) {
	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory, nil, nil,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)

	b := new(bytes.Buffer)
	err := serializer.Encode(obj, b)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
