package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func TestDefinition_GenerateSchemaForType(t *testing.T) {
	type MenuItem struct {
		ItemName string `json:"item_name" description:"Menu item name" required:"true"`
		Quantity int    `json:"quantity" description:"Quantity of menu item ordered" required:"true"`
		Price    int    `json:"price" description:"Price of the menu item" required:"true"`
	}

	type UserOrder struct {
		MenuItems       []MenuItem `json:"menu_items" description:"List of menu items ordered by the user" required:"true"`
		DeliveryAddress string     `json:"delivery_address" description:"Delivery address for the order" required:"true"`
		UserName        string     `json:"user_name" description:"Name of the user placing the order" required:"true"`
		PhoneNumber     string     `json:"phone_number" description:"Phone number of the user" required:"true"`
		PaymentMethod   string     `json:"payment_method" description:"Payment method" required:"true" enum:"cash,transfer"`
	}

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "Test MenuItem Schema",
			input: MenuItem{},
			want: `{
   "type":"object",
   "additionalProperties":false,
   "properties":{
      "item_name":{
         "type":"string",
         "description":"Menu item name"
      },
      "quantity":{
         "type":"integer",
         "description":"Quantity of menu item ordered"
      },
      "price":{
         "type":"integer",
         "description":"Price of the menu item"
      }
   },
   "required":[
      "item_name",
      "quantity",
      "price"
   ]
}`,
		},
		{
			name:  "Test UserOrder Schema",
			input: UserOrder{},
			want: `{
   "type":"object",
   "additionalProperties":false,
   "properties":{
      "menu_items":{
         "type":"array",
         "description":"List of menu items ordered by the user",
         "items":{
            "type":"object",
            "additionalProperties":false,
            "properties":{
               "item_name":{
                  "type":"string",
                  "description":"Menu item name"
               },
               "quantity":{
                  "type":"integer",
                  "description":"Quantity of menu item ordered"
               },
               "price":{
                  "type":"integer",
                  "description":"Price of the menu item"
               }
            },
            "required":[
               "item_name",
               "quantity",
               "price"
            ]
         }
      },
      "delivery_address":{
         "type":"string",
         "description":"Delivery address for the order"
      },
      "user_name":{
         "type":"string",
         "description":"Name of the user placing the order"
      },
      "phone_number":{
         "type":"string",
         "description":"Phone number of the user"
      },
      "payment_method":{
         "type":"string",
         "description":"Payment method",
         "enum":[
            "cash",
            "transfer"
         ]
      }
   },
   "required":[
      "menu_items",
      "delivery_address",
      "user_name",
      "phone_number",
      "payment_method"
   ]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate schema
			got, err := jsonschema.GenerateSchemaForType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSchemaForType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Convert both the generated schema and the expected JSON to maps for comparison
			wantBytes := []byte(tt.want)
			var want map[string]interface{}
			err = json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Errorf("Failed to Unmarshal expected JSON: error = %v", err)
				return
			}

			gotMap := structToMap(t, got)

			// Compare the maps
			if !reflect.DeepEqual(gotMap, want) {
				t.Errorf("GenerateSchemaForType() got = %v, want %v", gotMap, want)
			}
		})
	}
}

func TestDefinition_SchemaGenerationComparison(t *testing.T) {
	type MenuItem struct {
		ItemName string  `json:"item_name" description:"Menu item name" required:"true"`
		Quantity int     `json:"quantity" description:"Quantity of menu item ordered" required:"true"`
		Price    float64 `json:"price" description:"Price of the menu item" required:"true"`
	}

	type UserOrder struct {
		MenuItems       []MenuItem `json:"menu_items" description:"List of menu items ordered by the user" required:"true"`
		DeliveryAddress string     `json:"delivery_address" description:"Delivery address for the order" required:"true"`
		UserName        string     `json:"user_name" description:"Name of the user placing the order" required:"true"`
		PhoneNumber     string     `json:"phone_number" description:"Phone number of the user" required:"true"`
		PaymentMethod   string     `json:"payment_method" description:"Payment method" required:"true" enum:"cash,transfer"`
	}

	// Manually created schema to compare against struct-generated schema
	manualSchema := &jsonschema.Definition{
		Type:                 jsonschema.Object,
		AdditionalProperties: false,
		Properties: map[string]jsonschema.Definition{
			"menu_items": {
				Type:        jsonschema.Array,
				Description: "List of menu items ordered by the user",
				Items: &jsonschema.Definition{
					Type:                 jsonschema.Object,
					AdditionalProperties: false,
					Properties: map[string]jsonschema.Definition{
						"item_name": {
							Type:        jsonschema.String,
							Description: "Menu item name",
						},
						"quantity": {
							Type:        jsonschema.Integer,
							Description: "Quantity of menu item ordered",
						},
						"price": {
							Type:        jsonschema.Number,
							Description: "Price of the menu item",
						},
					},
					Required: []string{"item_name", "quantity", "price"},
				},
			},
			"delivery_address": {
				Type:        jsonschema.String,
				Description: "Delivery address for the order",
			},
			"user_name": {
				Type:        jsonschema.String,
				Description: "Name of the user placing the order",
			},
			"phone_number": {
				Type:        jsonschema.String,
				Description: "Phone number of the user",
			},
			"payment_method": {
				Type:        jsonschema.String,
				Description: "Payment method",
				Enum:        []string{"cash", "transfer"},
			},
		},
		Required: []string{"menu_items", "delivery_address", "user_name", "phone_number", "payment_method"},
	}

	t.Run("Compare Struct-Generated and Manual Schema", func(t *testing.T) {
		// Generate schema from struct
		structSchema, err := jsonschema.GenerateSchemaForType(UserOrder{})
		if err != nil {
			t.Fatalf("Failed to generate schema from struct: %v", err)
		}

		// Convert both schemas to maps for comparison
		structMap := structToMap(t, structSchema)
		manualMap := structToMap(t, manualSchema)

		// Compare the maps
		if !reflect.DeepEqual(structMap, manualMap) {
			t.Errorf("Schema generated from struct and manual schema do not match")
			t.Errorf("Struct generated schema: %v", structMap)
			t.Errorf("Manual schema: %v", manualMap)
		}
	})
}

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
				Type:        jsonschema.String,
				Description: "A string type",
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: jsonschema.String,
					},
				},
			},
			want: `{
   "type":"string",
   "description":"A string type",
   "properties":{
      "name":{
         "type":"string"
      }
   }
}`,
		},
		{
			name: "Test with nested Definition properties",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: jsonschema.String,
							},
							"age": {
								Type: jsonschema.Integer,
							},
						},
					},
				},
			},
			want: `{
   "type":"object",
   "properties":{
      "user":{
         "type":"object",
         "properties":{
            "name":{
               "type":"string"
            },
            "age":{
               "type":"integer"
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with complex nested Definition",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: jsonschema.String,
							},
							"age": {
								Type: jsonschema.Integer,
							},
							"address": {
								Type: jsonschema.Object,
								Properties: map[string]jsonschema.Definition{
									"city": {
										Type: jsonschema.String,
									},
									"country": {
										Type: jsonschema.String,
									},
								},
							},
						},
					},
				},
			},
			want: `{
   "type":"object",
   "properties":{
      "user":{
         "type":"object",
         "properties":{
            "name":{
               "type":"string"
            },
            "age":{
               "type":"integer"
            },
            "address":{
               "type":"object",
               "properties":{
                  "city":{
                     "type":"string"
                  },
                  "country":{
                     "type":"string"
                  }
               }
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with Array type Definition",
			def: jsonschema.Definition{
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{
					Type: jsonschema.String,
				},
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: jsonschema.String,
					},
				},
			},
			want: `{
   "type":"array",
   "items":{
      "type":"string"
   },
   "properties":{
      "name":{
         "type":"string"
      }
   }
}`,
		},
		{
			name: "Test with Enum type Definition",
			def: jsonschema.Definition{
				Type: jsonschema.String,
				Enum: []string{"celsius", "fahrenheit"},
			},
			want: `{
			   "type":"string",
			   "enum":["celsius","fahrenheit"]
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
