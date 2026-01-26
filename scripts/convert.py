# ruff: noqa
"""
Simple script to extract all enums to a specific component.

https://github.com/OpenAPITools/openapi-generator/issues/9567
"""

import json
import sys
from typing import Any
import re
import hashlib


# input_file = "./._openapi-generator/v3/openapi-raw.json"
# output_file = "./._openapi-generator/v3/openapi.json"

input_file = sys.argv[1]
output_file = sys.argv[2]


cache: dict[str, str] = {}

# When multiple inline enums appear for the same logical property across different
# schemas (e.g. "framework"), we want to unify them into a single component enum
# and merge all values together so nothing is lost.
# key: property name → component schema name
UNIFY_ENUMS_BY_PROPERTY: dict[str, str] = {
    "framework": "ENUMPROPERTIESFRAMEWORK",
}

def md5(s: str) -> str:
    return hashlib.md5(s.encode()).hexdigest()


def _enum_name(path: list[str], value: dict) -> str:
    # If the enum belongs to a known property we want to unify, return a fixed name
    # like ENUMPROPERTIESFRAMEWORK regardless of the path to ensure merging.
    try:
        for i, p in enumerate(path):
            if p == "properties" and i + 1 < len(path):
                prop_name = path[i + 1]
                if prop_name in UNIFY_ENUMS_BY_PROPERTY:
                    return UNIFY_ENUMS_BY_PROPERTY[prop_name]
    except Exception:
        pass

    cache_key = '-'.join(sorted(value["enum"]))
    if cache_key in cache:
        return cache[cache_key]

    raw: str = "Enum_"
    if len(value["enum"]) == 1:
        raw += str(value["enum"][0])
    else:
        raw += "_".join(path)
    final = re.sub(r"[^a-zA-Z0-9_]", "_", raw).upper()
    if len(final) > 200:
        final =  "Enum_" + md5("-".join(sorted(value["enum"])))
    cache[cache_key] = final
    return final


def _rec_extract_enums(schema: dict, path: list[str]) -> tuple[dict, dict[str, dict]]:
    """Recursively:
    - Extract inline enums into components and replace with $ref (keep nullable on the property)
    - Filter object.required to drop fields that are marked optional (nullable: true)
    - Enforce custom required additions when needed
    """
    
    enums: dict[str, dict] = {}

    # If the current schema node itself is an enum, extract it immediately
    if isinstance(schema, dict) and 'enum' in schema:
        target = _enum_name(path, schema)
        # Merge enum values if the target already exists
        if target in enums and 'enum' in enums[target]:
            existing = set(enums[target]['enum'])
            incoming = set(schema['enum'])
            merged = sorted(existing.union(incoming))
            schema = dict(schema)
            schema['enum'] = merged
        enums[target] = schema
        ref_node_1: dict[str, Any] = {"$ref": "#/components/schemas/" + target}
        if schema.get('nullable', False):
            ref_node_1['nullable'] = True
        return ref_node_1, enums

    if isinstance(schema, dict):
        # If this is an object schema, first process its properties
        if isinstance(schema.get('properties'), dict):
            props: dict = schema['properties']
            new_props: dict = {}
            for prop_name, prop_schema in list(props.items()):
                new_prop_schema, new_enums = _rec_extract_enums(prop_schema, path + ['properties', prop_name])
                if new_enums:
                    enums.update(new_enums)
                new_props[prop_name] = new_prop_schema
            schema['properties'] = new_props

            # Filter required: drop any property marked nullable
            if isinstance(schema.get('required'), list):
                filtered_required = [
                    r for r in schema['required']
                    if r in new_props and not bool(new_props.get(r, {}).get('nullable', False))
                ]
                schema['required'] = filtered_required

            # Custom rule: if schema has asset_class property, ensure both are required
            if 'asset_class' in new_props:
                existing_required = list(schema.get('required', []))
                for must in ['asset_class', 'asset_type']:
                    if must not in existing_required:
                        existing_required.append(must)
                schema['required'] = existing_required

        # Process other keys within this schema
        for key, value in list(schema.items()):
            if key == 'properties':
                continue  # already handled above

            if isinstance(value, dict):
                if 'enum' in value:
                    target = _enum_name(path + [key], value)
                    # Merge enum values if same target already discovered in this subtree
                    if target in enums and 'enum' in enums[target]:
                        existing = set(enums[target]['enum'])
                        incoming = set(value['enum'])
                        merged = sorted(existing.union(incoming))
                        value = dict(value)
                        value['enum'] = merged
                    enums[target] = value
                    ref_schema: dict[str, Any] = {"$ref": "#/components/schemas/" + target}
                    if value.get('nullable', False):
                        ref_schema['nullable'] = True
                    schema[key] = ref_schema
                else:
                    new_schema, new_enums = _rec_extract_enums(value, path + [key])
                    if new_enums:
                        enums.update(new_enums)
                    schema[key] = new_schema
            elif isinstance(value, list):
                for i in range(len(value)):
                    vi = value[i]
                    if isinstance(vi, dict):
                        new_schema, new_enums = _rec_extract_enums(vi, path + [str(i)])
                        if new_enums:
                            enums.update(new_enums)
                        schema[key][i] = new_schema

    return schema, enums

with open(input_file, "r") as f:
    raw = (
        json.dumps(json.loads(f.read()))
        .replace("#/$defs/", "#/components/schemas/")
        .replace('"examples": [],', '')
    )
    consts = re.findall(r'"const": "([^"]+)",', raw)
    for const in consts:
        raw = raw.replace(f'"const": "{const}",', f'"enum": ["{const}"],')

    raw = re.compile(r'"anyOf": \[[^"]*"type": "([^"]+)"[^"]*\]', re.MULTILINE).sub(r'"type": "\1"', raw)
    raw = re.compile(r'"anyOf": \[[^"]*"\$ref": "([^"]+)"[^"]*\]', re.MULTILINE).sub(r'"$ref": "\1"', raw)
    data = json.loads(raw)

for path, path_data in data["paths"].items():
    for method, operation_object in path_data.items():
        if method not in [
            "get",
            "put",
            "post",
            "delete",
            "options",
            "head",
            "patch",
            "trace",
        ]:
            continue

        # Remove "Beta" tag from operations that have other tags to avoid duplicate type declarations
        # The generator creates request types per tag, causing redeclarations when the same operation
        # appears under multiple tags (e.g., both "Beta" and "Integrations")
        tags = operation_object.get("tags", [])
        if len(tags) > 1 and "Beta" in tags:
            operation_object["tags"] = [t for t in tags if t != "Beta"]
        
        list_params = [
            "profileIds",
            "assetIds",
            "domains",
            "ids",
            "scanIds",
            "issueIds",
            "stages",
            "attachments",
            "tagsIds",
            "tagIds",
            "search",
            "jiraTicket",
            "risks",
            "assetClasses",
            "scannerKinds",
            "severities",
            "status",
            "levels",
            "types",
            "statuses",
            "type",
            "initiator",
            "initiators",
            "kinds",
            "locationIds",
        ]

        # Normalize list-like query params
        # - For params backed by arrayOrSingle on the server (initiators, kinds, risks, locationIds),
        #   expose them as arrays so the client sends repeated query params (form+explode)
        # - Keep others as strings (server accepts CSV via commaSeparatedString)
        if "parameters" in operation_object:
            for param in operation_object["parameters"]:
                name = param.get("name")
                if name in list_params:
                    if name in ["initiators", "kinds", "risks", "locationIds"]:
                        param["schema"] = {"type": "array", "items": {"type": "string"}}
                        param["style"] = "form"
                        param["explode"] = True
                    else:
                        param["schema"] = {"type": "string"}
                    
        responses: dict[str, dict] = operation_object.get("responses", {})
        if not responses:
            continue
        if (
            json_schema1 := operation_object.get("requestBody", {})
            .get("content", {})
            .get("application/json", {})
            .get("schema", {})
        ):
            schema, enums = _rec_extract_enums(json_schema1, [])
            if enums:
                # Merge enums into global components, preserving and unifying values
                for name, enum_schema in enums.items():
                    if name in data["components"]["schemas"] and 'enum' in enum_schema:
                        existing_schema = data["components"]["schemas"][name]
                        if 'enum' in existing_schema:
                            merged = sorted(set(existing_schema['enum']).union(set(enum_schema['enum'])))
                            existing_schema['enum'] = merged
                            data["components"]["schemas"][name] = existing_schema
                        else:
                            data["components"]["schemas"][name] = enum_schema
                    else:
                        data["components"]["schemas"][name] = enum_schema
                data["paths"][path][method]["requestBody"]["content"]["application/json"]["schema"] = schema

        for status_code, response_object in responses.items():
            if "content" not in response_object:
                continue
            content = response_object.get("content", {})
            if "application/json" not in content:
                continue
            json_schema2: dict = content.get("application/json", {}).get("schema", {})
            schema, enums = _rec_extract_enums(json_schema2, [])
            if enums:
                for name, enum_schema in enums.items():
                    if name in data["components"]["schemas"] and 'enum' in enum_schema:
                        existing_schema = data["components"]["schemas"][name]
                        if 'enum' in existing_schema:
                            merged = sorted(set(existing_schema['enum']).union(set(enum_schema['enum'])))
                            existing_schema['enum'] = merged
                            data["components"]["schemas"][name] = existing_schema
                        else:
                            data["components"]["schemas"][name] = enum_schema
                    else:
                        data["components"]["schemas"][name] = enum_schema
                data["paths"][path][method]["responses"][status_code]["content"]["application/json"]["schema"] = schema

# Unify all integration list endpoints to return ListIntegrations200Response
# This ensures all integration types (akamai, kubernetes, aws, etc.) use the same response type
unified_integration_response_schema = None
for path, path_data in data["paths"].items():
    # Match GET endpoints under /integrations/{type} pattern (but not /integrations/{type}/{id})
    if path.startswith("/integrations/") and not path.endswith("/{id}") and "get" in path_data:
        operation = path_data["get"]
        responses = operation.get("responses", {})
        if "200" in responses:
            response_200 = responses["200"]
            content = response_200.get("content", {})
            if "application/json" in content:
                json_schema = content["application/json"].get("schema", {})
                if json_schema and "$ref" not in json_schema:
                    # Use the first integration response schema as the template
                    if unified_integration_response_schema is None:
                        # Deep copy the schema
                        unified_integration_response_schema = json.loads(json.dumps(json_schema))
                        # Process it to extract enums
                        processed_schema, enums = _rec_extract_enums(unified_integration_response_schema, [])
                        # Store extracted enums in components
                        for enum_name, enum_schema in enums.items():
                            if enum_name in data["components"]["schemas"] and 'enum' in enum_schema:
                                existing_schema = data["components"]["schemas"][enum_name]
                                if 'enum' in existing_schema:
                                    merged = sorted(set(existing_schema['enum']).union(set(enum_schema['enum'])))
                                    existing_schema['enum'] = merged
                                    data["components"]["schemas"][enum_name] = existing_schema
                                else:
                                    data["components"]["schemas"][enum_name] = enum_schema
                            else:
                                data["components"]["schemas"][enum_name] = enum_schema
                        unified_integration_response_schema = processed_schema
                    break

# If we found a unified schema, create the component and replace all integration list responses
if unified_integration_response_schema:
    data["components"]["schemas"]["ListIntegrations200Response"] = unified_integration_response_schema
    
    # Replace all integration list GET endpoints with the unified response
    for path, path_data in data["paths"].items():
        if path.startswith("/integrations/") and not path.endswith("/{id}") and "get" in path_data:
            operation = path_data["get"]
            responses = operation.get("responses", {})
            if "200" in responses:
                response_200 = responses["200"]
                content = response_200.get("content", {})
                if "application/json" in content:
                    content["application/json"]["schema"] = {"$ref": "#/components/schemas/ListIntegrations200Response"}

with open(output_file, "w") as f:
    json.dump(data, f, indent=2)