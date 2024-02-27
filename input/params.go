package input

import (
  "google.golang.org/protobuf/proto"
  "google.golang.org/protobuf/types/pluginpb"
  "io"
  "os"
  "strings"
)

type Params struct {
	EmitOnly []string
	NamespaceMap map[string]string
	CollapseFields []string
	RemoveEnumPrefixes bool
	PreserveNonStringMaps bool
}

func ReadRequest() (*pluginpb.CodeGeneratorRequest, error) {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(in, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func parseRawParams(req *pluginpb.CodeGeneratorRequest) map[string]string {
	param := req.GetParameter()
	if len(param) == 0 {
		return nil
	}
	paramTokens := strings.Split(param, ",")
	paramMap := map[string]string{}
	for _, token := range paramTokens {
		paramStrings := strings.Split(token, "=")
		if len(paramStrings) == 2 {
			paramMap[paramStrings[0]] = paramStrings[1]
		}
	}
	return paramMap
}

func ParseParams(req *pluginpb.CodeGeneratorRequest) Params {
	params := Params{NamespaceMap: map[string]string{}}
	rawParams := parseRawParams(req)
	for k, v := range rawParams {
		if k == "emit_only" {
			params.EmitOnly = strings.Split(v, ";")
		} else if k == "namespace_map" {
			namespaces := strings.Split(v, ";")
			for _, namespaceMapToken := range namespaces {
				namespaceTokens := strings.Split(namespaceMapToken, ":")
				params.NamespaceMap[namespaceTokens[0]] = namespaceTokens[1]
			}
		} else if k == "collapse_fields" {
			params.CollapseFields = strings.Split(v, ";")
		} else if k == "remove_enum_prefixes" {
			params.RemoveEnumPrefixes = v == "true"
		} else if k == "preserve_non_string_maps" {
			params.PreserveNonStringMaps = true
		}
	}
	return params
}
