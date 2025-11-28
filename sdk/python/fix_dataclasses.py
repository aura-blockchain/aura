#!/usr/bin/env python3
"""Fix all dataclass field ordering issues."""
import os
import re
import ast

def fix_dataclass_fields(file_path):
    """Fix dataclass field ordering in a Python file."""
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Try to parse and see if it has errors
    try:
        ast.parse(content)
        return False  # No errors
    except TypeError as e:
        if "non-default argument" not in str(e):
            return False

        print(f"Fixing {file_path}: {e}")

        # For now, just report the file - manual fix needed
        return True

def main():
    types_dir = 'aura/types'
    for filename in os.listdir(types_dir):
        if filename.endswith('.py') and filename != '__init__.py':
            file_path = os.path.join(types_dir, filename)
            try:
                fix_dataclass_fields(file_path)
            except Exception as e:
                print(f"Error processing {file_path}: {e}")

if __name__ == '__main__':
    main()
