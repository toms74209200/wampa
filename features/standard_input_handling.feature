Feature: Standard Input Handling
  As a user
  I want to provide content via standard input
  So that I can dynamically generate content and integrate with shell pipelines

  Background:
    Given Wampa is installed
    And I have a file named "sample.md" with content:
      """
      # Sample File Content
      This is existing content in a file.
      """

  Scenario: Reading from standard input only
    When I run "wampa -s -o output.md << EOF
    # Standard Input Content
    This content comes from standard input.
    EOF"
    Then the file "output.md" should contain:
      """
      [//]: # "filepath: stdin"
      # Standard Input Content
      This content comes from standard input.
      """
    And the exit code should be 0

  Scenario: Reading from standard input and file inputs
    When I run "wampa -s -i sample.md -o output.md << EOF
    # Standard Input Content
    This content comes from standard input.
    EOF"
    Then the file "output.md" should contain:
      """
      [//]: # "filepath: stdin"
      # Standard Input Content
      This content comes from standard input.

      [//]: # "filepath: sample.md"
      # Sample File Content
      This is existing content in a file.
      """
    And the exit code should be 0

  Scenario: Using pipe for standard input
    When I run "echo '# Piped Content' | wampa -s -i sample.md -o output.md"
    Then the file "output.md" should contain:
      """
      [//]: # "filepath: stdin"
      # Piped Content

      [//]: # "filepath: sample.md"
      # Sample File Content
      This is existing content in a file.
      """
    And the exit code should be 0

  Scenario: Using long form flag for standard input
    When I run "wampa --stdin -i sample.md -o output.md << EOF
    # Standard Input Content
    This content comes from standard input.
    EOF"
    Then the file "output.md" should contain:
      """
      [//]: # "filepath: stdin"
      # Standard Input Content
      This content comes from standard input.

      [//]: # "filepath: sample.md"
      # Sample File Content
      This is existing content in a file.
      """
    And the exit code should be 0

  Scenario: Reading from standard input with configuration file
    Given I have a file named "wampa.json" with content:
      """
      {
        "input_files": ["sample.md"],
        "output_file": "output.md"
      }
      """
    When I run "wampa -s << EOF
    # Standard Input Content
    This content comes from standard input.
    EOF"
    Then the file "output.md" should contain:
      """
      [//]: # "filepath: stdin"
      # Standard Input Content
      This content comes from standard input.

      [//]: # "filepath: sample.md"
      # Sample File Content
      This is existing content in a file.
      """
    And the exit code should be 0

  Scenario: Using pipe with configuration file
    Given I have a file named "wampa.json" with content:
      """
      {
        "input_files": ["sample.md"],
        "output_file": "output.md"
      }
      """
    When I run "curl -s https://raw.githubusercontent.com/toms74209200/wampa/028171afb7eefed15d055b4d82618280c9782f74/TODO.md | wampa -s"
    Then the file "output.md" should contain content from "stdin" and "sample.md"
    And the exit code should be 0

  Scenario: Standard input flag with no actual input provided
    When I run "wampa -s -i sample.md -o output.md"
    Then the stderr should contain "Error: standard input flag was specified but no content was provided"
    And the exit code should be 1