#!/usr/bin/env python3
"""
Script to automatically wrap unwrapped errors in Go code with contextual fmt.Errorf calls.
"""
import os
import re
import sys

def analyze_context(lines, line_idx):
    """Analyze the function context to generate meaningful error message."""
    # Find the function name
    func_name = ""
    for i in range(line_idx - 1, max(0, line_idx - 50), -1):
        match = re.match(r'func\s+(?:\([^)]+\)\s+)?(\w+)', lines[i])
        if match:
            func_name = match.group(1)
            break

    # Look at the preceding lines for context
    operation = ""
    entity = ""

    # Common patterns to extract context
    for i in range(max(0, line_idx - 10), line_idx):
        line = lines[i].strip()

        # Look for error checks with assignment
        if 'if err' in line:
            # Check the line before for the operation
            if i > 0:
                prev = lines[i-1].strip()
                # Extract operation from common patterns
                if ':=' in prev:
                    # Extract variable name or method call
                    parts = prev.split(':=')
                    if len(parts) > 1:
                        rhs = parts[1].strip()
                        # Extract method calls
                        if '.' in rhs:
                            method_match = re.search(r'\.(\w+)\(', rhs)
                            if method_match:
                                operation = method_match.group(1)
                        # Look for common operations
                        if 'Unmarshal' in rhs:
                            operation = 'unmarshal'
                        elif 'Marshal' in rhs:
                            operation = 'marshal'
                        elif 'Get' in rhs:
                            operation = 'get'
                        elif 'Set' in rhs:
                            operation = 'set'
                        elif 'Delete' in rhs:
                            operation = 'delete'
                        elif 'Iterator' in rhs:
                            operation = 'create iterator'

        # Look for variable names that might indicate entity
        if 'did' in line.lower() and not entity:
            entity_match = re.search(r'(\w+did\w*)', line, re.IGNORECASE)
            if entity_match:
                entity = f" for {entity_match.group(1)}"
        elif 'id' in line.lower() and not entity:
            entity_match = re.search(r'(\w+[Ii]d\w*)', line)
            if entity_match:
                entity = f" for {entity_match.group(1)}"

    # Generate error message
    if operation:
        msg = f"failed to {operation}{entity}"
    elif func_name:
        msg = f"error in {func_name}{entity}"
    else:
        msg = "operation failed"

    return msg

def wrap_error(file_path):
    """Process a single file and wrap unwrapped errors."""
    try:
        with open(file_path, 'r') as f:
            lines = f.readlines()
    except Exception as e:
        print(f"Error reading {file_path}: {e}")
        return 0

    modified = False
    changes_made = 0
    new_lines = []
    needs_fmt_import = False

    for i, line in enumerate(lines):
        # Check if this line is just "return err"
        if re.match(r'^\s+return err\s*$', line):
            # Get the context
            context_msg = analyze_context(lines, i)

            # Get the indentation
            indent = re.match(r'^(\s+)', line).group(1) if re.match(r'^(\s+)', line) else ''

            # Replace with wrapped error
            new_line = f'{indent}return fmt.Errorf("{context_msg}: %w", err)\n'
            new_lines.append(new_line)
            modified = True
            changes_made += 1
            needs_fmt_import = True
        else:
            new_lines.append(line)

    if not modified:
        return 0

    # Check if fmt is already imported
    has_fmt_import = False
    for line in new_lines:
        if re.search(r'^\s*"fmt"\s*$', line):
            has_fmt_import = True
            break

    # Add fmt import if needed
    if needs_fmt_import and not has_fmt_import:
        # Find the import block
        import_added = False
        final_lines = []
        in_import_block = False

        for i, line in enumerate(new_lines):
            if re.match(r'^\s*import\s*\(', line):
                in_import_block = True
                final_lines.append(line)
                # Add fmt as first import
                indent = '\t'
                final_lines.append(f'{indent}"fmt"\n')
                final_lines.append('\n')
                import_added = True
            elif re.match(r'^\s*import\s+"', line) and not import_added:
                # Single import line, add fmt import before it
                final_lines.append('import "fmt"\n')
                final_lines.append(line)
                import_added = True
            else:
                final_lines.append(line)

        new_lines = final_lines

    # Write the modified file
    try:
        with open(file_path, 'w') as f:
            f.writelines(new_lines)
        return changes_made
    except Exception as e:
        print(f"Error writing {file_path}: {e}")
        return 0

def main():
    base_dir = "/home/hudson/blockchain-projects/aura/chain"

    # Get list of all keeper files with unwrapped errors
    import subprocess
    result = subprocess.run(
        'find x/*/keeper -name "*.go" ! -name "*_test.go" -exec grep -l "return err$" {} +',
        shell=True,
        capture_output=True,
        text=True,
        cwd=base_dir
    )

    files = [f.strip() for f in result.stdout.strip().split('\n') if f.strip()]

    print(f"Processing {len(files)} files...")
    total_changes = 0

    for file_path in files:
        full_path = os.path.join(base_dir, file_path)
        changes = wrap_error(full_path)
        if changes > 0:
            print(f"✓ {file_path}: wrapped {changes} errors")
            total_changes += changes

    print(f"\nTotal: wrapped {total_changes} errors across {len(files)} files")

if __name__ == "__main__":
    main()
