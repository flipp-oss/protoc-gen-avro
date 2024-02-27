package avro

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	"google.golang.org/protobuf/types/descriptorpb"
	"slices"
)

type Record struct {
	Name string
	Namespace string
	Fields []Field
}

func (t Record) GetName() string {
	return t.Name
}

func (t Record) GetNamespace() string {
	return t.Namespace
}

func (t Record) ToJSON(types *TypeRepo) (any, error) {
	if slices.Contains(types.CollapseFields, t.Name) {
		return t.Fields[0].ToJSON(types)
	}
	types.SeenType(t)
	jsonMap := orderedmap.New()
	jsonMap.Set("type", "record")
	jsonMap.Set("name", t.Name)
	jsonMap.Set("namespace", types.MappedNamespace(t.Namespace))
	fields := make([]any, len(t.Fields))
	for i, field := range t.Fields {
		fieldJson, err := field.ToJSON(types)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s in record %s: %w", field.Name, t.Name, err)
		}
		fields[i] = fieldJson
	}
	jsonMap.Set("fields", fields)
	return jsonMap, nil
}

func RecordFromProto(proto *descriptorpb.DescriptorProto, namespace string, typeRepo *TypeRepo) []NamedType {
	// Protobuf `map` is actually a record with a `map_entry` boolean set, and two fields ("key" and "value").
	// Avro only supports maps with string keys, but Protobuf supports maps with any key type.
	// If we detect a string key, we can turn this into an Avro map, but otherwise we have to leave it as a record.
	if proto.Options.GetMapEntry() {
		if !typeRepo.PreserveNonStringMaps || proto.Field[0].GetType() == descriptorpb.FieldDescriptorProto_TYPE_STRING {
			return []NamedType{
				Map{
					Namespace: namespace,
					Name:      proto.GetName(),
					Values:    FieldTypeFromProto(proto.Field[1]),
				},
			}
		}
	}

	var fields []Field
	oneofs := make([]Field, len(proto.OneofDecl))
	nested := make([]NamedType, len(proto.NestedType))
//	enums := make([]Enum, len(proto.EnumType))
	for i, field := range proto.NestedType {
		nested[i] = RecordFromProto(field, fmt.Sprintf("%s.%s", namespace, proto.GetName()), typeRepo)[0]
	}
	//for i, field := range proto.EnumType {
	//	enums[i] = EnumFromProto(field)
	//}
	for i, oneof := range proto.OneofDecl {
		oneofs[i] = Field{
			Name: oneof.GetName(),
			Type: Union{Types: []Type{}},
		}
	}
	for _, field := range proto.Field {
		// proto3_optional will put in a dummy oneof. We don't need to emit this (all Protobuf values
		// are optional) so we should ignore it.
		if field.OneofIndex != nil && !field.GetProto3Optional() {
			union := oneofs[field.GetOneofIndex()].Type.(Union)
			union.Types = append(union.Types, FieldTypeFromProto(field))
			oneofs[field.GetOneofIndex()].Type = union
		} else {
			fields = append(fields, FieldFromProto(field))
		}
	}
	for _, oneof := range oneofs {
		union := oneof.Type.(Union)
		if len(union.Types) > 0 {
			fields = append(fields, oneof)
		}
	}
	//for _, enum := range enums {
	//	fields = append(fields, Field{
	//		Name: enum.Name,
	//		Type: enum,
	//		Default: enum.Symbols[0],
	//	})
	//}
	var types []NamedType
	for _, subRecord := range nested {
		types = append(types, subRecord)
	}
	types = append(types, Record{
		Name: proto.GetName(),
		Namespace: namespace,
		Fields: fields,
	})
	return types
}
