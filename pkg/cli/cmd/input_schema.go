package cmd

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unicode"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
)

const (
	// createMethodArgs is the reflect method arity of a generated create call:
	// receiver plus context.
	createMethodArgs = 2
	// updateMethodArgs is the reflect method arity of a generated update call:
	// receiver, context, and the resource id.
	updateMethodArgs = 3
	// setterMethodArgs is the reflect method arity of a generated body setter:
	// receiver plus the request struct.
	setterMethodArgs = 2

	jsonSchemaDraft = "https://json-schema.org/draft/2020-12/schema"
)

// additionalPropertiesFalse is shared by integration variants that have a
// generated model. A kind with no model leaves the keyword unset and says so.
var additionalPropertiesFalse = false

// providedSchema advertises a JSON Schema that is not the reflection of a
// single Go struct. CommandSchemaRegistry stores it as the command input.
type providedSchema struct {
	schema *clischema.JSONSchema
}

func (p providedSchema) JSONSchema() *clischema.JSONSchema { return p.schema }

func mustInputSchema(build func() (*clischema.JSONSchema, error)) providedSchema {
	schema, err := build()
	if err != nil {
		panic(err)
	}

	return providedSchema{schema: schema}
}

var (
	enumOnce         sync.Once
	enumValuesByType = map[reflect.Type][]string{}

	assetSchemaOnce sync.Once
	assetSchema     *clischema.JSONSchema
	assetSchemaErr  error

	integrationSchemaOnce sync.Once
	integrationCreateBody *clischema.JSONSchema
	integrationUpdateBody *clischema.JSONSchema
	integrationSchemaErr  error

	kindsOnce sync.Once
	kinds     []string
)

func ensureEnumValues() {
	enumOnce.Do(func() {
		registerEnums(
			v3.AllowedENUMAPKEnumValues,
			v3.AllowedENUMAWSLAMBDAEnumValues,
			v3.AllowedENUMBITBUCKETREPOSITORYEnumValues,
			v3.AllowedENUMDNSEnumValues,
			v3.AllowedENUMGITHUBREPOSITORYEnumValues,
			v3.AllowedENUMGITLABREPOSITORYEnumValues,
			v3.AllowedENUMGRAPHQLEnumValues,
			v3.AllowedENUMGRPCEnumValues,
			v3.AllowedENUMIPV4EnumValues,
			v3.AllowedENUMIPV4RANGEEnumValues,
			v3.AllowedENUMIPV6EnumValues,
			v3.AllowedENUMMCPEnumValues,
			v3.AllowedENUMPACKAGEEnumValues,
			v3.AllowedENUMRESTEnumValues,
			v3.AllowedENUMSCHEMAEnumValues,
			v3.AllowedENUMSOAPEnumValues,
			v3.AllowedENUMSOFTWAREEnumValues,
			v3.AllowedENUMWEBAPPEnumValues,
			v3.AllowedENUMWEBSOCKETEnumValues,
		)
	})
}

func registerEnums(slices ...any) {
	for _, slice := range slices {
		value := reflect.ValueOf(slice)
		if value.Kind() != reflect.Slice || value.Len() == 0 {
			continue
		}

		values := make([]string, value.Len())
		for i := 0; i < value.Len(); i++ {
			values[i] = value.Index(i).String()
		}

		enumValuesByType[value.Type().Elem()] = values
	}
}

func assetCreateInputSchema() (*clischema.JSONSchema, error) {
	assetSchemaOnce.Do(func() {
		assetSchema, assetSchemaErr = buildAssetCreateInputSchema()
	})

	return assetSchema, assetSchemaErr
}

func integrationCreateInputSchema() (*clischema.JSONSchema, error) {
	if err := loadIntegrationSchemas(); err != nil {
		return nil, err
	}

	return integrationCreateBody, nil
}

func integrationUpdateInputSchema() (*clischema.JSONSchema, error) {
	if err := loadIntegrationSchemas(); err != nil {
		return nil, err
	}

	return integrationUpdateBody, nil
}

func loadIntegrationSchemas() error {
	integrationSchemaOnce.Do(func() {
		integrationCreateBody, integrationSchemaErr = buildIntegrationInputSchema("Create", createMethodArgs)
		if integrationSchemaErr != nil {
			return
		}

		integrationUpdateBody, integrationSchemaErr = buildIntegrationInputSchema("Update", updateMethodArgs)
	})

	return integrationSchemaErr
}

// integrationKinds returns the --kind path segments the generated integration
// client can call, derived from IntegrationsAPIService method names.
func integrationKinds() []string {
	kindsOnce.Do(func() {
		ops, err := integrationOps("Create", createMethodArgs)
		if err != nil {
			panic(err)
		}

		kinds = make([]string, 0, len(ops))
		for _, op := range ops {
			kinds = append(kinds, op.kind)
		}
	})

	return append([]string(nil), kinds...)
}

func buildAssetCreateInputSchema() (*clischema.JSONSchema, error) {
	ensureEnumValues()
	payloads, err := assetCreatePayloads()
	if err != nil {
		return nil, err
	}

	variants := make([]*clischema.JSONSchema, 0, len(payloads))
	seen := map[string]struct{}{}
	for _, payload := range payloads {
		for _, concrete := range schemaTypesFor(payload) {
			variant, err := assetVariant(concrete)
			if err != nil {
				return nil, err
			}

			key := variant.Properties["asset_type"].Enum[0] + " " + variant.Description
			if _, ok := seen[key]; ok {
				return nil, fmt.Errorf("duplicate asset create variant %s", key)
			}

			seen[key] = struct{}{}
			variants = append(variants, variant)
		}
	}

	sort.Slice(variants, func(i, j int) bool {
		return variants[i].Description < variants[j].Description
	})

	return &clischema.JSONSchema{
		Schema: jsonSchemaDraft,
		Description: "Create-asset body. oneOf lists every asset type the CLI dispatches on " +
			"asset_type. Use " + climcp.GetToolSpecToolName + " for the full input schema.",
		OneOf: variants,
	}, nil
}

func assetVariant(payload reflect.Type) (*clischema.JSONSchema, error) {
	variant := generatedObject(zeroValue(payload))
	values, err := assetTypeEnum(payload)
	if err != nil {
		return nil, err
	}

	prop := variant.Properties["asset_type"]
	if prop == nil {
		return nil, fmt.Errorf("%s schema has no asset_type property", payload.Name())
	}

	// Generate may have hoisted this field. The concrete asset_type list is
	// the one the dispatcher matches, so it stays inline on the property.
	prop.Ref = ""
	if prop.Type == nil {
		prop.Type = "string"
	}

	prop.Enum = values
	variant.Description = fmt.Sprintf("%s. asset_type: %s.", payload.Name(), strings.Join(values, ", "))
	variant.Schema = ""

	return variant, nil
}

func assetTypeEnum(payload reflect.Type) ([]string, error) {
	if payload.Kind() == reflect.Pointer {
		payload = payload.Elem()
	}

	field, ok := payload.FieldByName("AssetType")
	if !ok {
		return nil, fmt.Errorf("%s has no AssetType field", payload.Name())
	}

	enumType := field.Type
	if enumType.Kind() == reflect.Pointer {
		enumType = enumType.Elem()
	}

	values := enumValuesByType[enumType]
	if len(values) == 0 {
		return nil, fmt.Errorf("no enum values registered for %s", enumType)
	}

	return append([]string(nil), values...), nil
}

func buildIntegrationInputSchema(prefix string, methodArgs int) (*clischema.JSONSchema, error) {
	ops, err := integrationOps(prefix, methodArgs)
	if err != nil {
		return nil, err
	}

	variants := make([]*clischema.JSONSchema, 0, len(ops))
	for _, op := range ops {
		variants = append(variants, integrationVariant(op))
	}

	// anyOf, not oneOf: kinds share fields such as name and parameters, and the
	// body does not carry kind (the --kind flag selects the API route). A valid
	// body therefore matches more than one variant. oneOf would reject it.
	// The description says that outright. A top-level anyOf that switched on
	// --kind would be rejected by MCP clients.
	unconstrained := make([]string, 0)
	for _, op := range ops {
		if op.payload == nil {
			unconstrained = append(unconstrained, op.kind)
		}
	}

	note := ""
	if len(unconstrained) > 0 {
		note = fmt.Sprintf(
			" Kinds with no generated model (%s) accept any object.",
			strings.Join(unconstrained, ", "),
		)
	}

	return &clischema.JSONSchema{
		Schema: jsonSchemaDraft,
		Type:   "object",
		Description: fmt.Sprintf(
			"JSON body forwarded to the route selected by --kind. --kind does not select a variant: anyOf lists each kind's fields, and a body can match more than one.%s Call %s for the full schema.",
			note,
			climcp.GetToolSpecToolName,
		),
		AnyOf: variants,
	}, nil
}

type integrationOp struct {
	kind    string
	payload reflect.Type
}

// integrationVariant is the body schema for one kind. A nil payload means the
// generated client has no usable request model for that kind (postman's
// setter is a copy of cloudflare's). The kind stays in the enum, and the body
// stays an object instead of advertising the wrong model.
func integrationVariant(op integrationOp) *clischema.JSONSchema {
	if op.payload == nil {
		return &clischema.JSONSchema{
			Type: "object",
			Description: fmt.Sprintf(
				"JSON body for --kind %q. The generated client has no request model for this kind, so this variant accepts any object.",
				op.kind,
			),
		}
	}

	variant := generatedObject(zeroValue(op.payload))
	required := "(none)"
	if len(variant.Required) > 0 {
		required = strings.Join(variant.Required, ", ")
	}

	variant.Description = fmt.Sprintf("JSON body for --kind %q. Required fields: %s.", op.kind, required)
	variant.Schema = ""
	variant.AdditionalProperties = &additionalPropertiesFalse

	return variant
}

func integrationOps(prefix string, methodArgs int) ([]integrationOp, error) {
	typ := reflect.TypeOf((*v3.IntegrationsAPIService)(nil))
	ops := make([]integrationOp, 0)
	seen := map[string]struct{}{}
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if !integrationMethod(method, prefix, methodArgs) {
			continue
		}

		payload, _ := setterPayload(method)
		kind := integrationKind(method.Name, prefix)
		if _, ok := seen[kind]; ok {
			return nil, fmt.Errorf("duplicate integration kind %q", kind)
		}

		seen[kind] = struct{}{}
		ops = append(ops, integrationOp{kind: kind, payload: payload})
	}

	if len(ops) == 0 {
		return nil, fmt.Errorf("no %s integration methods", prefix)
	}

	sort.Slice(ops, func(i, j int) bool { return ops[i].kind < ops[j].kind })

	return ops, nil
}

func integrationMethod(method reflect.Method, prefix string, methodArgs int) bool {
	name := method.Name
	if strings.HasSuffix(name, "Execute") || !strings.HasPrefix(name, prefix) {
		return false
	}

	if !strings.HasSuffix(name, "Integration") {
		return false
	}

	return method.Type.NumIn() == methodArgs
}

func integrationKind(methodName, prefix string) string {
	rest := strings.TrimPrefix(methodName, prefix)
	rest = strings.TrimSuffix(rest, "Integration")

	return camelToKebab(rest)
}

func camelToKebab(value string) string {
	var b strings.Builder
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('-')
			}

			b.WriteRune(unicode.ToLower(r))
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

// assetCreatePayloads returns the request types CreateAsset can dispatch to.
// Discovery follows AssetsAPIService the same way CreateAsset does: Create*
// methods that take only a context and have a typed body setter. Comment
// creation is excluded because it takes an asset id.
func assetCreatePayloads() ([]reflect.Type, error) {
	typ := reflect.TypeOf((*v3.AssetsAPIService)(nil))
	payloads := make([]reflect.Type, 0)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		name := method.Name
		if strings.HasSuffix(name, "Execute") || !strings.HasPrefix(name, "Create") {
			continue
		}

		if method.Type.NumIn() != createMethodArgs {
			continue
		}

		payload, ok := setterPayload(method)
		if !ok {
			continue
		}

		payloads = append(payloads, payload)
	}

	if len(payloads) == 0 {
		return nil, errors.New("no asset create methods")
	}

	sort.Slice(payloads, func(i, j int) bool { return payloads[i].Name() < payloads[j].Name() })

	return payloads, nil
}

func setterPayload(method reflect.Method) (reflect.Type, bool) {
	if method.Type.NumOut() == 0 {
		return nil, false
	}

	ret := method.Type.Out(0)
	base := ret
	if base.Kind() == reflect.Pointer {
		base = base.Elem()
	}

	setterName := method.Name + "Request"
	setter, ok := base.MethodByName(setterName)
	if !ok {
		setter, ok = reflect.PointerTo(base).MethodByName(setterName)
	}

	if !ok || setter.Type.NumIn() != setterMethodArgs {
		return nil, false
	}

	return setter.Type.In(1), true
}

// schemaTypesFor unwraps an OpenAPI anyOf wrapper (a struct whose exported
// fields are all untagged pointers to the real request models). Every other
// type is returned as itself.
func schemaTypesFor(payload reflect.Type) []reflect.Type {
	t := payload
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return []reflect.Type{t}
	}

	inners := make([]reflect.Type, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Tag.Get("json") != "" || field.Type.Kind() != reflect.Pointer || field.Type.Elem().Kind() != reflect.Struct {
			return []reflect.Type{t}
		}

		inners = append(inners, field.Type.Elem())
	}

	if len(inners) == 0 {
		return []reflect.Type{t}
	}

	return inners
}

func generatedObject(v any) *clischema.JSONSchema {
	schema := clischema.Generate(v)
	schema.Schema = ""

	return schema
}

func zeroValue(t reflect.Type) any {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return reflect.Zero(t).Interface()
}
