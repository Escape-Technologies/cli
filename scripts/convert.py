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

def md5(s: str) -> str:
    return hashlib.md5(s.encode()).hexdigest()


def _enum_name(path: list[str], value: dict) -> str:
    if '-'.join(sorted(value["enum"])) in cache:
        return cache['-'.join(sorted(value["enum"]))]
    raw: str = "Enum_"
    if len(value["enum"]) == 1:
        raw += str(value["enum"][0])
    else:
        raw += "_".join(path)
    final = re.sub(r"[^a-zA-Z0-9_]", "_", raw).upper()
    if len(final) > 200:
        final =  "Enum_" + md5("-".join(sorted(value["enum"])))
    cache['-'.join(sorted(value["enum"]))] = final
    return final


def _rec_extract_enums(schema: dict, path: list[str]) -> tuple[dict, dict[str, dict]]:
    """Recursively:
    - Extract inline enums into components and replace with $ref (keep nullable on the property)
    - Filter object.required to drop fields that are marked optional (nullable: true)
    - Enforce custom required additions when needed
    """
    enums: dict[str, dict] = {}

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
                    enums[target] = value
                    ref_node: dict[str, Any] = {"$ref": "#/components/schemas/" + target}
                    if value.get('nullable', False):
                        ref_node['nullable'] = True
                    schema[key] = ref_node
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
                data["components"]["schemas"].update(enums)
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
                data["components"]["schemas"].update(enums)
                data["paths"][path][method]["responses"][status_code]["content"]["application/json"]["schema"] = schema

with open(output_file, "w") as f:
    json.dump(data, f, indent=2)
