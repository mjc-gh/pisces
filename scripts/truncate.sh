#!/bin/bash

jq '
  # Truncate strings to 32 chars at top level
  .certificate_info.sans |= .[0:4] |
  
  # Process top-level values (truncate strings to 32)
  with_entries(
    if .value | type == "string" then
      .value |= .[0:32]
    else . end
  ) |
  
  # Process assets array
  .assets |= (
    map(
      # Truncate string values in each asset
      with_entries(
        if .value | type == "string" then
          .value |= .[0:32]
        else . end
      ) |
      # Truncate certificate_info.sans if it exists
      if .certificate_info then
        .certificate_info.sans |= .[0:4]
      else . end
    ) |
    # Remove duplicates and "Other" resource_types
    reduce .[] as $item (
      {seen: [], result: []};
      if $item.resource_type == "Other" then
        .
      elif (.seen | index($item.resource_type)) then
        .
      else
        .seen += [$item.resource_type] |
        .result += [$item]
      end
    ) | .result
  ) |
  
  # Final walk for any remaining strings > 64
  walk(if type == "string" and length > 64 then .[0:64] + "..." else . end)
' "$@"
