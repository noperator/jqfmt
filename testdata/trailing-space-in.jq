map(select(has("resource")) | .resource.github_repository | to_entries | map(.value | map(.name))) | flatten[]
