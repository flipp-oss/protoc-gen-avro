package main

import (
    "fmt"
    "github.com/stretchr/testify/assert"
    "os"
  "os/exec"
    "path/filepath"
    "strings"
  "testing"
)

// Overall approach taken from https://github.com/mix-php/mix/blob/master/src/grpc/protoc-gen-mix/plugin_test.go

// When the environment variable RUN_AS_PROTOC_GEN_AVRO is set, we skip running
// tests and instead act as protoc-gen-avro. This allows the test binary to
// pass itself to protoc.
func init() {
    if os.Getenv("RUN_AS_PROTOC_GEN_AVRO") != "" {
        main()
        os.Exit(0)
    }
}

func fileNames(directory string, appendDirectory bool) ([]string, error) {
    files, err := os.ReadDir(directory)
    if err != nil {
        return nil, fmt.Errorf("can't read %s directory: %w", directory, err)
    }
    var names []string
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        if appendDirectory {
          names = append(names, filepath.Base(directory) + "/" + file.Name())
        } else {
            names = append(names, file.Name())
        }
    }
    return names, nil
}

func runTest(t *testing.T, directory string, options map[string]string) {
    workdir, _ := os.Getwd()
    tmpdir, err := os.MkdirTemp(workdir, "proto-test.")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpdir)

    args := []string{
        "-I.",
        "--avro_out=" + tmpdir,
    }
    names, err := fileNames(workdir + "/testdata", true)
    if err != nil {
        t.Fatal(fmt.Errorf("testData fileNames %w", err))
    }
    for _, name := range names {
        args = append(args, name)
    }
    for k, v := range options {
        args = append(args, "--avro_opt=" + k + "=" + v)
    }
    protoc(t, args)

    testDir := workdir + "/testdata/" + directory
    if os.Getenv("UPDATE_SNAPSHOTS") != "" {
        cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cp %v/* %v", tmpdir, testDir))
        cmd.Run()
    } else {
        assertEqualFiles(t, testDir, tmpdir)
    }
}

func Test_Base(t *testing.T) {
    runTest(t, "base", map[string]string{})
}

func Test_CollapseFields(t *testing.T) {
    runTest(t, "collapse_fields", map[string]string{"collapse_fields": "StringList"})
}

func Test_EmitOnly(t *testing.T) {
    runTest(t, "emit_only", map[string]string{"emit_only": "Widget"})
}

func Test_NamespaceMap(t *testing.T) {
    runTest(t, "namespace_map", map[string]string{"namespace_map": "testdata:mynamespace"})
}

func Test_PreserveNonStringMaps(t *testing.T) {
    runTest(t, "preserve_non_string_maps", map[string]string{"preserve_non_string_maps": "true"})
}

func assertEqualFiles(t *testing.T, original, generated string) {
    names, err := fileNames(original, false)
    if err != nil {
        t.Fatal(fmt.Errorf("original fileNames %w", err))
    }
    generatedNames, err := fileNames(generated, false)
    if err != nil {
        t.Fatal(fmt.Errorf("generated fileNames %w", err))
    }
    assert.Equal(t, names, generatedNames)
    for i, name := range names {
        originalData, err := os.ReadFile(original + "/" + name)
        if err != nil {
            t.Fatal("Can't find original file for comparison")
        }

        generatedData, err := os.ReadFile(generated + "/" + generatedNames[i])
        if err != nil {
            t.Fatal("Can't find generated file for comparison")
        }
        r := strings.NewReplacer("\r\n", "", "\n", "")
        assert.Equal(t, r.Replace(string(originalData)), r.Replace(string(generatedData)))
    }
}

func protoc(t *testing.T, args []string) {
    cmd := exec.Command("protoc", "--plugin=protoc-gen-avro=" + os.Args[0])
    cmd.Args = append(cmd.Args, args...)
    cmd.Env = append(os.Environ(), "RUN_AS_PROTOC_GEN_AVRO=1")
    out, err := cmd.CombinedOutput()

    if len(out) > 0 || err != nil {
        t.Log("RUNNING: ", strings.Join(cmd.Args, " "))
    }

    if len(out) > 0 {
        t.Log(string(out))
    }

    if err != nil {
        t.Fatalf("protoc: %v", err)
    }
}
