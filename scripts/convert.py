#! /usr/bin/env python3
"""
Simple script to extract all enums to a specific component.

https://github.com/OpenAPITools/openapi-generator/issues/9567
"""

import json
import sys
import hashlib


def md5hash(input: str) -> str:
    return hashlib.md5(input.encode()).hexdigest()


def _rec_extract_enums(schema: dict, path: list[str]) -> tuple[dict, dict[str, dict]]:
    enums = {}
    for key, value in list(schema.items()):
        if isinstance(value, dict):
            if "enum" in value:
                target = "Enum_" + (
                    str(value["enum"][0]).upper().replace(" ", "_").replace(".", "_")
                    if len(value["enum"]) == 1
                    else md5hash(str(value["enum"]))
                )
                enums[target] = value
                schema[key] = {"$ref": "#/components/schemas/" + target}
            else:
                new_schema, new_enums = _rec_extract_enums(value, path + [key])
                if new_enums:
                    enums.update(new_enums)
                    schema[key] = new_schema
        elif isinstance(value, list):
            for i in list(range(len(value))):
                if isinstance(value[i], dict):
                    new_schema, new_enums = _rec_extract_enums(
                        value[i], path + [str(i)]
                    )
                    if new_enums:
                        enums.update(new_enums)
                        schema[key][i] = new_schema
    return schema, enums


if len(sys.argv) != 3:
    print("Usage: convert.py <input_file> <output_file>")
    sys.exit(1)

input_file = sys.argv[1]
output_file = sys.argv[2]

with open(input_file, "r") as f:
    data = json.load(f)

if "components" not in data:
    data["components"] = {}

if "schemas" not in data["components"]:
    data["components"]["schemas"] = {}

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
        for status_code, response_object in responses.items():
            if "content" not in response_object:
                continue
            content = response_object.get("content", {})
            if "application/json" not in content:
                continue
            json_schema: dict = content.get("application/json", {}).get("schema", {})
            schema, enums = _rec_extract_enums(json_schema, [])
            if enums:
                data["components"]["schemas"].update(enums)
                data["paths"][path][method]["responses"][status_code]["content"][
                    "application/json"
                ]["schema"] = schema

with open(output_file, "w") as f:
    json.dump(data, f, indent=2)
