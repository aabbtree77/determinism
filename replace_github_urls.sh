#!/bin/bash

# Replace occurrences of the old GitHub URL with the new one in all files
replace_in_files() {
    find . -type f -print0 | while IFS= read -r -d '' file; do
        sed -i 's|https://github.com/aabbtree77/determinism|https://github.com/aabbtree77/determinism|g' "$file"
    done
}

# Execute replacement
replace_in_files

echo "Replacement of URLs completed."

