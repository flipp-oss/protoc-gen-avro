package main

import (
	"encoding/json"
	"fmt"
	"github.com/flipp-oss/protoc-gen-avro/avro"
	"github.com/flipp-oss/protoc-gen-avro/input"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"log"
	"os"
	"slices"
)

var params input.Params
var typeRepo *avro.TypeRepo

func processMessage(proto *descriptorpb.DescriptorProto, protoPackage string) {
	records := avro.RecordFromProto(proto, protoPackage, typeRepo)
	for _, record := range records {
		typeRepo.AddType(record)
	}
}

func processEnum(proto *descriptorpb.EnumDescriptorProto, protoPackage string) {
	enum := avro.EnumFromProto(proto, protoPackage)
	typeRepo.AddType(enum)
}

func generateFileResponse(record avro.Record) (*pluginpb.CodeGeneratorResponse_File, error) {
	typeRepo.Start()
	fileName := fmt.Sprintf("%s.avsc", record.Name)
	jsonObj, err := record.ToJSON(typeRepo)
	if err != nil {
		return nil, fmt.Errorf("error parsing record %s: %w", record.Name, err)
	}
	jsonBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}
	jsonString := string(jsonBytes)
	return &pluginpb.CodeGeneratorResponse_File{
		Name:    &fileName,
		Content: &jsonString,
	}, nil
}

func generateResponse(recordsToEmit []string) *pluginpb.CodeGeneratorResponse {
	feature := uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	response := &pluginpb.CodeGeneratorResponse{
		SupportedFeatures: &feature,
	}
	var records []avro.Type
	if recordsToEmit != nil {
		for _, recordName := range recordsToEmit {
			record := typeRepo.GetTypeByBareName(recordName)
			if record == nil {
				errString := fmt.Errorf("record %s not found", recordName).Error()
				response.Error = &errString
			} else {
				records = append(records, record)
			}
		}
	} else {
		for _, record := range typeRepo.Types {
			records = append(records, record)
		}
	}
	for _, t := range records {
		record, ok := t.(avro.Record)
		if ok {
			file, err := generateFileResponse(record)
			if err != nil {
				errString := fmt.Errorf("error getting JSON for record %s: %w", record.Name, err).Error()
				response.Error = &errString
				return response
			}
			response.File = append(response.File, file)
		}
	}
	return response
}

func processAll(fileProto *descriptorpb.FileDescriptorProto) {
	for _, t := range fileProto.MessageType {
		processMessage(t, fileProto.GetPackage())
	}
	for _, t := range fileProto.EnumType {
		processEnum(t, fileProto.GetPackage())
	}
}

func writeResponse(req *pluginpb.CodeGeneratorRequest) {
	response := generateResponse(params.EmitOnly)
	out, err := proto.Marshal(response)
	if err != nil {
		log.Fatalf("%s", fmt.Errorf("error marshalling response: %w", err))
	}
	_, err = os.Stdout.Write(out)
	if err != nil {
		log.Fatalf("%s", fmt.Errorf("error writing response: %w", err))
	}
}

func main() {
	req, err := input.ReadRequest()
	if err != nil {
		log.Fatalf("%s", fmt.Errorf("error reading request: %w", err))
	}
	params = input.ParseParams(req)
	typeRepo = avro.NewTypeRepo(params)

	for _, file := range req.ProtoFile {
		if !slices.Contains(req.FileToGenerate, *file.Name) {
			continue
		}
		processAll(file)
	}
	writeResponse(req)
}
