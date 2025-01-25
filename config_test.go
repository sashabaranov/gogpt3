package openai_test

import (
	"testing"

	"github.com/alejandrojnm/go-openai"
)

func TestGetAzureDeploymentByModel(t *testing.T) {
	cases := []struct {
		Model                string
		AzureModelMapperFunc func(model string) string
		Expect               string
	}{
		{
			Model:  "gpt-3.5-turbo",
			Expect: "gpt-35-turbo",
		},
		{
			Model:  "gpt-3.5-turbo-0301",
			Expect: "gpt-35-turbo-0301",
		},
		{
			Model:  "text-embedding-ada-002",
			Expect: "text-embedding-ada-002",
		},
		{
			Model:  "",
			Expect: "",
		},
		{
			Model:  "models",
			Expect: "models",
		},
		{
			Model:  "gpt-3.5-turbo",
			Expect: "my-gpt35",
			AzureModelMapperFunc: func(model string) string {
				modelmapper := map[string]string{
					"gpt-3.5-turbo": "my-gpt35",
				}
				if val, ok := modelmapper[model]; ok {
					return val
				}
				return model
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Model, func(t *testing.T) {
			conf := openai.DefaultAzureConfig("", "https://test.openai.azure.com/")
			if c.AzureModelMapperFunc != nil {
				conf.AzureModelMapperFunc = c.AzureModelMapperFunc
			}
			actual := conf.GetAzureDeploymentByModel(c.Model)
			if actual != c.Expect {
				t.Errorf("Expected %s, got %s", c.Expect, actual)
			}
		})
	}
}
