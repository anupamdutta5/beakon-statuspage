#!/usr/bin/env python3
"""
Script to automatically refactor c.JSON error responses to use RespondWithError helper
"""

import re

# Read the file
with open('internal/handlers/tenant_admin_handler.go', 'r') as f:
    content = f.read()

# Pattern to match error responses with gin.H
# Matches: c.JSON(http.StatusXXX, gin.H{"error": "message"})
pattern = r'c\.JSON\((http\.Status(?:BadRequest|InternalServerError|NotFound|Unauthorized|ServiceUnavailable|NotImplemented)),\s*gin\.H\{"error":\s*(".*?")\}\)'

# Replacement: RespondWithError(c, http.StatusXXX, "message")
def replace_func(match):
    status = match.group(1)
    message = match.group(2)
    return f'RespondWithError(c, {status}, {message})'

# Perform replacement
new_content = re.sub(pattern, replace_func, content)

# Count replacements
original_count = len(re.findall(pattern, content))
print(f"Found {original_count} error responses to refactor")

# Write back
with open('internal/handlers/tenant_admin_handler.go', 'w') as f:
    f.write(new_content)

print(f"Refactored {original_count} error responses")
print("Done!")
