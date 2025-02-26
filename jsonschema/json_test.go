package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestDefinition_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		def  jsonschema.Definition
		want string
	}{
		{
			name: "Test with empty Definition",
			def:  jsonschema.Definition{},
			want: `{}`,
		},
		{
			name: "Test with Definition properties set",
			def: jsonschema.Definition{
				Type:        []jsonschema.DataType{jsonschema.String},
				Description: "A string type",
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: []jsonschema.DataType{jsonschema.String},
					},
				},
			},
			want: `{
   "type":["string"],
   "description":"A string type",
   "properties":{
      "name":{
         "type":["string"]
      }
   }
}`,
		},
		{
			name: "Test with nested Definition properties",
			def: jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Object},
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: []jsonschema.DataType{jsonschema.Object},
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: []jsonschema.DataType{jsonschema.String},
							},
							"age": {
								Type: []jsonschema.DataType{jsonschema.Integer},
							},
						},
					},
				},
			},
			want: `{
   "type":["object"],
   "properties":{
      "user":{
         "type":["object"],
         "properties":{
            "name":{
               "type":["string"]
            },
            "age":{
               "type":["integer"]
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with complex nested Definition",
			def: jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Object},
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: []jsonschema.DataType{jsonschema.Object},
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: []jsonschema.DataType{jsonschema.String},
							},
							"age": {
								Type: []jsonschema.DataType{jsonschema.Integer},
							},
							"address": {
								Type: []jsonschema.DataType{jsonschema.Object},
								Properties: map[string]jsonschema.Definition{
									"city": {
										Type: []jsonschema.DataType{jsonschema.String},
									},
									"country": {
										Type: []jsonschema.DataType{jsonschema.String},
									},
								},
							},
						},
					},
				},
			},
			want: `{
   "type":["object"],
   "properties":{
      "user":{
         "type":["object"],
         "properties":{
            "address":{
               "type":["object"],
               "properties":{
                  "city":{
                     "type":["string"]
                  },
                  "country":{
                     "type":["string"]
                  }
               }
            },
			"name":{
               "type":["string"]
            },
            "age":{
               "type":["integer"]
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with Array type Definition",
			def: jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Array},
				Items: &jsonschema.Definition{
					Type: []jsonschema.DataType{jsonschema.String},
				},
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: []jsonschema.DataType{jsonschema.String},
					},
				},
			},
			want: `{
   "type":["array"],
   "items":{
      "type":["string"]
   },
   "properties":{
      "name":{
         "type":["string"]
      }
   }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantBytes := []byte(tt.want)
			var want map[string]interface{}
			err := json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}

			got := structToMap(t, tt.def)
			gotPtr := structToMap(t, &tt.def)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, want)
			}
			if !reflect.DeepEqual(gotPtr, want) {
				t.Errorf("MarshalJSON() gotPtr = %v, want %v", gotPtr, want)
			}
		})
	}
}

func structToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	gotBytes, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Failed to Marshal JSON: error = %v", err)
		return nil
	}

	var got map[string]interface{}
	err = json.Unmarshal(gotBytes, &got)
	if err != nil {
		t.Errorf("Failed to Unmarshal JSON: error =  %v", err)
		return nil
	}
	return got
}

type basicStruct struct {
	Name string `json:"name"`
}

type nestedStruct struct {
	User basicStruct `json:"user"`
}

type nullableValueStruct struct {
	Name *string `json:"name"`
}

type arrayStruct struct {
	Names []string `json:"names"`
}

func TestGenerateSchemaForType(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		args    args
		want    *jsonschema.Definition
		wantErr bool
	}{
		{
			name: "Test with basic struct",
			args: args{v: basicStruct{}},
			want: &jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Object},
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: []jsonschema.DataType{jsonschema.String},
					},
				},
				Required:             []string{"name"},
				AdditionalProperties: false,
			},
		},
		{
			name: "Test with nested struct",
			args: args{v: nestedStruct{}},
			want: &jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Object},
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: []jsonschema.DataType{jsonschema.Object},
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: []jsonschema.DataType{jsonschema.String},
							},
						},
						Required:             []string{"name"},
						AdditionalProperties: false,
					},
				},
				Required:             []string{"user"},
				AdditionalProperties: false,
			},
		},
		{
			name: "Test with nullable value struct",
			args: args{v: nullableValueStruct{}},
			want: &jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Object},
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: []jsonschema.DataType{jsonschema.String, jsonschema.Null},
					},
				},
				Required:             []string{"name"},
				AdditionalProperties: false,
			},
		},
		{
			name: "Test with array struct",
			args: args{v: arrayStruct{}},
			want: &jsonschema.Definition{
				Type: []jsonschema.DataType{jsonschema.Object},
				Properties: map[string]jsonschema.Definition{
					"names": {
						Type: []jsonschema.DataType{jsonschema.Array},
						Items: &jsonschema.Definition{
							Type: []jsonschema.DataType{jsonschema.String},
						},
					},
				},
				Required:             []string{"names"},
				AdditionalProperties: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonschema.GenerateSchemaForType(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSchemaForType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSchemaForType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
