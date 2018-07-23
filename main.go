package main

import (
	"io/ioutil"
	"os"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/ti/protoc-gen-rest/rest"
	"strings"

	"bufio"
)

func main() {
	// Begin by allocating a generator. The request and response structures are stored there
	// so we can do error handling easily - the response structure contains the field to
	// report failure.
	g := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	g.CommandLineParameters(g.Request.GetParameter())

	// Create a wrapped version of the Descriptors and EnumDescriptors that
	// point to the file that defines them.
	g.WrapTypes()

	g.SetPackageNames()
	g.BuildTypeNameMap()

	g.GenerateAllFiles()

	for i, _ := range g.Response.File {
		fileName := *g.Response.File[i].Name
		index := strings.LastIndex(fileName, ".")
		rename := fileName
		if len(fileName) > index {
			rename = fileName[:index] + "." + plugin.Name + fileName[index:]
		}
		g.Response.File[i].Name = &rename
		content := *g.Response.File[i].Content
		p0, p1 := strings.Index(content,`import`),  strings.Index(content,`import (`)
		var genImports string
		imports := content[p0:p1]
		scanner := bufio.NewScanner(strings.NewReader(imports))
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Index(text," proto ") > 0 {
				continue
			}
			if strings.Index(text,".") > 0 {
				genImports += "\n" + scanner.Text()
			}
		}
		genImports += "\n\n"

		pkgImports := content[p1: strings.Index(content,`// Reference imports`)]

		firstInex := strings.Index(content,`//Start Services`) + 16
		lastInex :=  strings.Index(content,`//End Services`)
		contentEnd :=content[0:p0]  + genImports + pkgImports +  content[firstInex:lastInex]
		contentEnd = strings.Replace(contentEnd,"protoc-gen-go","protoc-gen-rest", -1)
		g.Response.File[i].Content = &contentEnd
	}

	// Send back the results.
	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}

