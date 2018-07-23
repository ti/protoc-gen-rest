package rest

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "golang.org/x/net/context"
	//Name the plugin name
	Name = "rest"
)

func init() {
	generator.RegisterPlugin(new(rest))
}

// rest is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for go-rest support.
type rest struct {
	gen *generator.Generator
}

func New()  *rest {
	return new(rest)
}
// Name returns the name of this plugin, "rest".
func (g *rest) Name() string {
	return Name
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	clientPkg  string
	serverPkg  string
)

// Init initializes the plugin.
func (g *rest) Init(gen *generator.Generator) {
	g.gen = gen
	contextPkg = generator.RegisterUniquePackageName("context", nil)
	clientPkg = generator.RegisterUniquePackageName("client", nil)
	serverPkg = generator.RegisterUniquePackageName("server", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *rest) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *rest) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *rest) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *rest) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("// Reference imports")
	g.P("var _ ", contextPkg, ".Context")
	g.P("var _ ", clientPkg, ".Option")
	g.P()

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *rest) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P("runtime", " ", strconv.Quote(path.Join(g.gen.ImportPrefix, "github.com/grpc-ecosystem/grpc-gateway/runtime")))
	g.P("grpc", " ", strconv.Quote(path.Join(g.gen.ImportPrefix, "google.golang.org/grpc")))
	g.P(contextPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, contextPkgPath)))
	g.P(")")
	g.P()
	g.P()
}


func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }


// generateService generates all the code for the named service.
func (g *rest) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	g.P("//Start Services")
	origServName := service.GetName()
	servName := generator.CamelCase(origServName)
	g.P("// Server API for ", servName, " service")
	// Server interface.
	serverType := unexport(servName) + "ServerClient"
	g.P("type ", serverType, " struct {")
	g.P("srv ", servName, "Server")
	g.P("}")
	g.P()
	// Server registration.
	g.P("func Register", servName, "ServerHandlerClient(ctx context.Context, mux *runtime.ServeMux, srv ", servName, "Server) error {")
	g.P("return Register", servName, "HandlerClient(ctx, mux,", "New", servName ,"ServerClient(srv))")
	g.P("}")
	g.P()

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Method {
		hname := g.generateServerMethod(servName, method)
		handlerNames = append(handlerNames, hname)
	}
	// Server registration.
	g.P("func New", servName, "ServerClient(srv ", servName, "Server)", servName, "Client{")
	g.P("return &", unexport(servName), "ServerClient{")
	g.P("srv:srv,")
	g.P("}")
	g.P("}")
	g.P()
	g.P("//End Services")
}

func (g *rest) generateServerMethod(servName string, method *pb.MethodDescriptorProto) string {
	methName := generator.CamelCase(method.GetName())
	hname := fmt.Sprintf("_%s_%s_Handler", servName + "Server", methName)
	inType := g.typeName(method.GetInputType())
	outType := g.typeName(method.GetOutputType())
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		g.P("func (h *",  unexport(servName) + "ServerClient", ") ", methName, "(ctx ", contextPkg, ".Context, in *", inType, ", _ ...grpc.CallOption", ") (*",outType, ", error) {")
		g.P("return h.", "srv", ".", methName, "(ctx, in)")
		g.P("}")
		g.P()
		return hname
	}
	outClientType := fmt.Sprintf("%s_%sClient", servName, methName)
	if method.GetServerStreaming() && !method.GetClientStreaming() {
		g.P("func (h *",  unexport(servName) + "ServerClient", ") ", methName, "(ctx ", contextPkg, ".Context, in *", inType, ", _ ...grpc.CallOption", ") (cli ",outClientType, ", err error) {")
		g.P("return")
		g.P("}")
		g.P()
	}
	if method.GetClientStreaming() {
		g.P("func (h *",  unexport(servName) + "ServerClient", ") ", methName, "(ctx ", contextPkg, ".Context", ", _ ...grpc.CallOption", ") (cli ",outClientType, ", err error) {")
		g.P("return")
		g.P("}")
		g.P()
	}
	return hname
}

// generateClientSignature returns the client-side signature for a method.
func (g *rest) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	reqArg := ", in *" + g.typeName(method.GetInputType())
	if method.GetClientStreaming() {
		reqArg = ""
	}
	respName := "*" + g.typeName(method.GetOutputType())
	if method.GetServerStreaming() || method.GetClientStreaming() {
		respName = servName + "_" + generator.CamelCase(origMethName) +  "serverClient"
	}

	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) (%s, error)", methName, contextPkg, reqArg, "grpc", respName)
}

// generateServerSignature returns the server-side signature for a method.
func (g *rest) generateServerSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	var reqArgs []string
	ret := "error"
	reqArgs = append(reqArgs, contextPkg+".Context")

	if !method.GetClientStreaming() {
		reqArgs = append(reqArgs, "*"+g.typeName(method.GetInputType()))
	}
	if method.GetServerStreaming() || method.GetClientStreaming() {
		reqArgs = append(reqArgs, servName+"_"+generator.CamelCase(origMethName)+"Stream")
	}
	if !method.GetClientStreaming() && !method.GetServerStreaming() {
		reqArgs = append(reqArgs, "*"+g.typeName(method.GetOutputType()))
	}
	return methName  + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}
