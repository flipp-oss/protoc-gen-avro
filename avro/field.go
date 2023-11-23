package avro

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Field struct {
	Name string
	Type Type
	Default any
}

func (t Field) ToJSON(types *TypeRepo) (any, error) {
	typeJson, err := t.Type.ToJSON(types)
	if err != nil {
		return nil, fmt.Errorf("error parsing field type %s %w", t.Name, err)
	}
  jsonMap := orderedmap.New()
  jsonMap.Set("name", t.Name)
  jsonMap.Set("type", typeJson)
  if t.Default != "" {
    jsonMap.Set("default", t.Default)
  } else {
		jsonMap.Set("default", DefaultValue(t.Type))
	}
  return jsonMap, nil
}

func FieldFromProto(proto *descriptorpb.FieldDescriptorProto) Field {
  return Field{
    Name: proto.GetName(),
    Type: FieldTypeFromProto(proto),
    Default: proto.GetDefaultValue(),
  }
}

func FieldTypeFromProto(proto *descriptorpb.FieldDescriptorProto) Type {
	basicType := BasicFieldTypeFromProto(proto)
	if proto.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		return Array{Items: basicType}
	} else if proto.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL {
		return Union{Types: []Type{Bare("null"), basicType}}
	} else {
		return basicType
	}
}

func BasicFieldTypeFromProto(proto *descriptorpb.FieldDescriptorProto) Type {
	switch proto.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return Bare("float")
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return Bare("double")
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return Bare("long")
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return Bare("int")
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return Bare("boolean")
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return Bare("string")
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return Ref(proto.GetTypeName())
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return Bare("bytes")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return Ref(proto.GetTypeName())
	}
	return Bare(proto.GetName())
}
