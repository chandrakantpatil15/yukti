#!/bin/bash

OUTPUT_FILE="all.md"

# Clear the output file
> "$OUTPUT_FILE"

# Loop through ONLY .md files in the current folder
shopt -s nullglob
for file in *.md; do
    # Skip the output file itself
    if [ "$file" == "$OUTPUT_FILE" ]; then
        continue
    fi

    echo "===== File: $file =====" >> "$OUTPUT_FILE"
    cat "$file" >> "$OUTPUT_FILE"
    echo -e "\n\n" >> "$OUTPUT_FILE"
done

echo "Combined all .md files into $OUTPUT_FILE"

