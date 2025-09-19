#!/bin/bash
# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# It reads from the single stages.yaml file and mimics the two-step
# lookup process of the original run.sh script.
get_value() {
    local key="$1"
    # If the key looks like a directory path, it means we need to find the
    # corresponding tfvars_path for it.
    if [[ "$key" == *"/"* || "$key" == "0"* ]]; then
        # Use yq to search for the entry where dir_path matches the key, then return the tfvars_path
        yq ".stages | to_entries | .[] | select(.value.dir_path == \"$key\") | .value.tfvars_path" config/stages.yaml
    # Otherwise, the key is a friendly name, and we need to find the dir_path.
    else
        # Use yq to look up the entry by its key and return the dir_path
        yq ".stages.\"$key\".dir_path" config/stages.yaml
    fi
}

# Export the function, making it available to any sub-processes.
# This allows it to override the function in the original run.sh script.
export -f get_value

# Dynamically find the directory where this wrapper script is located.
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Define the path to the real run.sh relative to THIS script's location.
REAL_RUN_SH_PATH="$SCRIPT_DIR/../../../../execution/run.sh"

# Get the directory of the real run.sh script.
REAL_RUN_SH_DIR=$(dirname "$REAL_RUN_SH_PATH")

# Change to the real script's directory. This is the crucial step.
cd "$REAL_RUN_SH_DIR" || exit

# Execute the script using just its filename, as we are in its directory.
exec "./$(basename "$REAL_RUN_SH_PATH")" "$@"