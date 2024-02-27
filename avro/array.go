package avro

import (
  "fmt"
  "github.com/iancoleman/orderedmap"
)

type Array struct {
  Items Type
}

func (t Array) ToJSON(types *TypeRepo) (any, error) {
  itemJson, err := t.Items.ToJSON(types)
  if err != nil {
    return nil, fmt.Errorf("error parsing item type: %w", err)
  }
 // you can't have a repeated map in protobuf, so if we are an array enclosing a map,
  // we are *really* just a plain map. We'd see this code path because a map is encoded
  // as a repeated "fake message type" that has key and value fields.
  mapType, ok := itemJson.(*orderedmap.OrderedMap)
  if ok {
    returnedType, _ := mapType.Get("type")
    if returnedType == "map" {
      return itemJson, nil
    }
    returnedMap, ok := returnedType.(*orderedmap.OrderedMap)
    if ok {
     _, hasType := returnedMap.Get("type")
     if hasType {
       // it's adding an extra nesting of type: { type: ... } } - we have to flatten it
       for _, k := range returnedMap.Keys() {
         val, _ := returnedMap.Get(k)
          mapType.Set(k, val)
       }
     }
    }
  }
  jsonMap := orderedmap.New()
  jsonMap.Set("type", "array")
  jsonMap.Set("items", itemJson)
  return jsonMap, nil
}
