# README: Unit Test Suite for `run.sh`

This directory contains a Go-based unit test suite for the main `run.sh` shell script.

The primary goal of this test suite is to verify the logic of `run.sh` in a fast, reliable, and automated way. It uses a **configuration-driven** approach where a **single `stages.yaml` file** defines all stage details and the entire test plan. The suite uses a **mock `terraform` executable**, which allows it to test the script's command generation logic without needing cloud credentials or running real Terraform commands.

***
## Key Tests Performed
The suite performs several critical checks:

* **`TestStaticAnalysis`**: Runs the `shellcheck` linter against `run.sh` to enforce a high standard of code quality and catch common shell scripting bugs.

* **`TestConfigurationSync`**: Verifies that the `valid_stages` variable hardcoded inside `run.sh` is perfectly synchronized with the master list of stages defined as keys in `config/stages.yaml`. This prevents "configuration drift."

* **`TestLogicAndCommandVerification`**: This is the core logic test, which verifies that `run.sh` generates the correct `terraform` commands. It reads the `test_plan` from `stages.yaml` to:
    * Run a **default set of commands** (e.g., `apply`, `init-apply`) for all standard stages.
    * Run a **stage-specific list of commands** that override the defaults.
    * Run **completely custom, one-off test cases** for unique scenarios.

***
## Test Suite Architecture
The suite is composed of a Go test file, a shell wrapper, and a single, comprehensive YAML configuration file.



### Go Code (`run_test.go`)
* **`setupTerraformMock`**: Creates a fake `terraform` executable on the fly that records the arguments it was called with, allowing the test to verify them.
* **`loadTestConfig`**: A single helper function that reads and parses the entire `config/stages.yaml` file into memory.
* **`generateTestCases`**: The "brain" of the test suite. It reads the `TestConfig` object and programmatically builds the final, comprehensive list of tests to run based on the `test_plan`.

### Shell Wrapper (`test_wrapper.sh`)
This script acts as a harness that allows the Go test to override the `get_value` function inside the `run.sh` script. It uses the `yq` command-line tool to read all mappings directly from the `config/stages.yaml` file.

### The Single Configuration File (`config/stages.yaml`)
All test data and configuration have been consolidated into this one file, which has three main sections:

| Section               | Purpose                                                                                                                                      |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| **`stages`** | The **master list** of all stages. It maps each stage's friendly name to its `dir_path` and `tfvars_path`.                                      |
| **`test_plan`** | Defines **which tests to run**. It contains defaults for standard stages, specific overrides, and completely custom one-off test cases.         |
| **`command_templates`**| Defines the expected output format for each Terraform command. This makes the Go test engine completely generic.                               |

***
## Extending the Test Suite 🚀
Adding or changing tests is now a simple, configuration-only process.

### Scenario 1: Adding a New Stage (with Default Tests)
This is the simplest case. The test suite will automatically run the `default_commands` from the `test_plan` for the new stage.
1.  In `config/stages.yaml`, **add a new entry** to the `stages:` map with the new stage's `dir_path` and `tfvars_path`.
2.  **Crucially, add the new stage's friendly name** to the `valid_stages` variable inside the original **`run.sh`** script to avoid "Invalid stage" errors.

That's it! No other changes are needed.

### Scenario 2: Customizing Tests for a Stage
To run a specific list of commands for a stage (either new or existing) instead of the defaults:
1.  In `config/stages.yaml`, add the stage's friendly name and a list of commands to the **`stage_specific_commands:`** map inside the `test_plan`.

### Scenario 3: Adding a Completely Custom Test
For a unique test that doesn't fit the standard pattern (e.g., testing a special combination of flags):
1.  In `config/stages.yaml`, add a new entry to the **`custom_test_cases:`** list inside the `test_plan`.

***
## Prerequisites
To run this test suite, your environment must have:
* Go (1.18+ recommended)
* `shellcheck`
* `yq` (v4+)