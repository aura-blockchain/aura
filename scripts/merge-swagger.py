#!/usr/bin/env python3
"""Merge individual Swagger/OpenAPI 2.0 specs into a unified spec."""

import json
import os
import sys
from pathlib import Path

def merge_swagger_files(input_dir: str, output_file: str):
    """Merge all swagger files into a single unified spec."""

    merged = {
        "swagger": "2.0",
        "info": {
            "title": "Aura Blockchain API",
            "description": "REST API for the Aura identity and privacy blockchain",
            "version": "1.0.0",
            "contact": {
                "name": "Aura Team"
            }
        },
        "host": "localhost:1317",
        "basePath": "/",
        "schemes": ["http", "https"],
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "tags": [],
        "paths": {},
        "definitions": {}
    }

    seen_tags = set()
    files_processed = 0

    for root, dirs, files in os.walk(input_dir):
        for file in files:
            if file.endswith('.swagger.json'):
                filepath = os.path.join(root, file)
                try:
                    with open(filepath, 'r') as f:
                        spec = json.load(f)

                    # Merge paths
                    for path, methods in spec.get('paths', {}).items():
                        if path not in merged['paths']:
                            merged['paths'][path] = methods
                        else:
                            for method, config in methods.items():
                                if method not in merged['paths'][path]:
                                    merged['paths'][path][method] = config

                    # Merge definitions
                    for name, schema in spec.get('definitions', {}).items():
                        if name not in merged['definitions']:
                            merged['definitions'][name] = schema

                    # Collect tags
                    for tag in spec.get('tags', []):
                        if tag['name'] not in seen_tags:
                            seen_tags.add(tag['name'])
                            merged['tags'].append(tag)

                    files_processed += 1
                except Exception as e:
                    print(f"Warning: Could not process {filepath}: {e}", file=sys.stderr)

    # Sort tags
    merged['tags'].sort(key=lambda x: x['name'])

    # Sort paths
    merged['paths'] = dict(sorted(merged['paths'].items()))

    # Write output
    with open(output_file, 'w') as f:
        json.dump(merged, f, indent=2)

    print(f"Merged {files_processed} files into {output_file}")
    print(f"Total paths: {len(merged['paths'])}")
    print(f"Total definitions: {len(merged['definitions'])}")
    print(f"Total tags: {len(merged['tags'])}")

if __name__ == '__main__':
    script_dir = Path(__file__).parent.parent
    input_dir = script_dir / 'docs' / 'api' / 'swagger'
    output_file = script_dir / 'docs' / 'api' / 'openapi.json'

    merge_swagger_files(str(input_dir), str(output_file))
