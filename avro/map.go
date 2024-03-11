package avro

import (
  "fmt"
  "github.com/iancoleman/orderedmap"
)

type Map struct {
  Name string
  Namespace string
  Values Type
}

func (t Map) GetName() string {
	return t.Name
}

func (t Map) GetNamespace() string {
	return t.Namespace
}

func (t Map) ToJSON(types *TypeRepo) (any, error) {
  valueJson, err := t.Values.ToJSON(types)
  if err != nil {
    return nil, fmt.Errorf("error parsing map value type: %w", err)
  }
  // values are pretty much never a union of null and something else - but since Protobuf is
  // optional by default, it seems like it is
  mapType, ok := valueJson.([]any)
  if ok && mapType[0] == "null"{
    valueJson = mapType[1]
  }
  valueJsonMap, ok := valueJson.(*orderedmap.OrderedMap)
  if ok {
    returnedType, _ := valueJsonMap.Get("type")
    returnedMap, ok := returnedType.(*orderedmap.OrderedMap)
    if ok {
     _, hasType := returnedMap.Get("type")
     if hasType {
       // it's adding an extra nesting of type: { type: ... } } - we have to flatten it
       for _, k := range returnedMap.Keys() {
         val, _ := returnedMap.Get(k)
          valueJsonMap.Set(k, val)
       }
     }
    }

  }
  jsonMap := orderedmap.New()
  jsonMap.Set("type", "map")
  jsonMap.Set("values", valueJson)
  return jsonMap, nil
}
