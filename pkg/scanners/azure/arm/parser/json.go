package parser

import (
	"github.com/aquasecurity/defsec/pkg/scanners/azure/arm/parser/armjson"
)

type Template struct {
	Schema         Value                `json:"$schema"`
	ContentVersion Value                `json:"contentVersion"`
	APIProfile     Value                `json:"apiProfile"`
	Parameters     map[string]Parameter `json:"parameters"`
	Variables      map[string]Value     `json:"variables"`
	Functions      []Function           `json:"functions"`
	Resources      []Resource           `json:"resources"`
	Outputs        map[string]Value     `json:"outputs"`
}

type Metadata struct {
	StartLine int
	EndLine   int
	Comments  []string
}

type Value struct {
	Metadata
	Raw  interface{}
	Type Type
}

type Parameter struct {
	Metadata
	Type         Value `json:"type"`
	DefaultValue Value `json:"defaultValue"`
	MaxLength    Value `json:"maxLength"`
	MinLength    Value `json:"minLength"`
}

type Type string

const (
	TypeBoolean Type = "boolean"
	TypeString  Type = "string"
	TypeNumber  Type = "number"
	TypeObject  Type = "object"
	TypeNull    Type = "null"
	TypeArray   Type = "array"
)

type Function struct{}

type innerMapValue map[string]Value

type MapValue struct {
	Metadata
	innerMapValue
}

type innerSliceValue []Value

type SliceValue struct {
	Metadata
	innerSliceValue
}

type Resource struct {
	Metadata `json:"-"`
	innerResource
}

type innerResource struct {
	APIVersion Value    `json:"apiVersion"`
	Type       Value    `json:"type"`
	Kind       Value    `json:"kind"`
	Name       Value    `json:"name"`
	Location   Value    `json:"location"`
	Tags       MapValue `json:"tags"`
	Sku        MapValue `json:"sku"`
	Properties MapValue `json:"properties"`
}

func (v *MapValue) UnmarshalJSONWithMetadata(node armjson.Node) error {
	tempMap := make(map[string]interface{})
	if err := node.Decode(&tempMap); err != nil {
		return err
	}

	if err := node.Decode(&v.innerMapValue); err != nil {
		return err
	}

	for _, val := range tempMap {
		switch t := val.(type) {
		case []interface{}:
			var sliceValues innerSliceValue
			for _, i := range t {
				sliceValues = append(sliceValues, Value{
					Metadata: v.Metadata, Raw: i,
				})
			}
			_ = SliceValue{Metadata: v.Metadata, innerSliceValue: sliceValues}
		}
	}

	if err := node.Decode(&v.innerMapValue); err != nil {
		return err
	}
	v.StartLine = node.Range().Start.Line
	v.EndLine = node.Range().End.Line
	for _, comment := range node.Comments() {
		var str string
		if err := comment.Decode(&str); err != nil {
			return err
		}
		v.Comments = append(v.Comments, str)
	}
	return nil
}

func (v *SliceValue) UnmarshalJSONWithMetadata(node armjson.Node) error {
	if err := node.Decode(&v.innerSliceValue); err != nil {
		return err
	}
	v.StartLine = node.Range().Start.Line
	v.EndLine = node.Range().End.Line
	for _, comment := range node.Comments() {
		var str string
		if err := comment.Decode(&str); err != nil {
			return err
		}
		v.Comments = append(v.Comments, str)
	}
	return nil
}

func (v *Resource) UnmarshalJSONWithMetadata(node armjson.Node) error {

	if err := node.Decode(&v.innerResource); err != nil {
		return err
	}

	v.Metadata = Metadata{
		StartLine: node.Range().Start.Line,
		EndLine:   node.Range().End.Line,
	}

	for _, comment := range node.Comments() {
		var str string
		if err := comment.Decode(&str); err != nil {
			return err
		}
		v.Metadata.Comments = append(v.Metadata.Comments, str)
	}

	return nil
}

func (v *Value) UnmarshalJSONWithMetadata(node armjson.Node) error {

	if err := node.Decode(&v.Raw); err != nil {
		return err
	}

	switch node.Kind() {
	case armjson.KindString:
		v.Type = TypeString
	case armjson.KindNumber:
		v.Type = TypeNumber
	case armjson.KindBoolean:
		v.Type = TypeBoolean
	case armjson.KindObject:
		v.Type = TypeObject
	case armjson.KindNull:
		v.Type = TypeNull
	case armjson.KindArray:
		v.Type = TypeArray
	default:
		panic(node.Kind())
	}

	v.StartLine = node.Range().Start.Line
	v.EndLine = node.Range().End.Line

	for _, comment := range node.Comments() {
		var str string
		if err := comment.Decode(&str); err != nil {
			return err
		}
		v.Comments = append(v.Comments, str)
	}
	return nil
}
