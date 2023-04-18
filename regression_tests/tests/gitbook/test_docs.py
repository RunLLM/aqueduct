import argparse
import os
import re
import subprocess
import time
from glob import glob
from pathlib import Path
from typing import List, NamedTuple

start = time.time()


class SnippetExec(NamedTuple):
    # The result from executing the code snippets sequentially.
    # We produce one per file.
    # Returned from `run` and stores useful information for debugging that is displayed if a file fails to run to the end successfully.

    # All the Python code block in the Markdown file joined in order of appearance.
    snippet: str
    success: bool
    # The output logged to stdout when running the snippet
    output: str
    # The error message.
    error: str


# Not in-scope for testing now
blacklist_sections = [
    "integrations",  # Requires setting up the integrations which currently cannot be done through the SDK alone
    "installation-and-configuration/installing-aqueduct",  # Specific examples that relate to screenshots of the external setups (e.g. using the IP address displayed on the screenshot)
    "api-reference",  # Header snippets.
    "example-workflows",  # Already tested elsewhere
]

blacklist_files = [
    "operators/configuring-resource-constraints.md",  # Header snippets. The GPU Access one require the operator to be executed on Kubernetes
    # "workflows/deleting-a-workflow.md",  # Referencing workflow UUIDs.
    "operators/specifying-a-requirements.txt.md",  # Requires specific requirements.txt
]

blacklist_snippets = {}


def get_code(page: str) -> List[str]:
    """
    Use an regular expression to find all Python code blocks in the page.
    Return
    - List of each code block found in the page in order of appearance
    """
    contents = Path(page).read_text()

    """
    The regular expression is used to extract code blocks written in the Python language that are enclosed in a triple backtick (```) fence.

    - `{3} matches exactly three occurrences of the previous character or group, which in this case is the backtick (`).
    - python\n matches the string "python" followed by a newline character.
    - ([\s\S]*?) is a capture group that matches any sequence of characters, including whitespace characters and line breaks. The *? quantifier means "match zero or more of the preceding token, but as few as possible".
    - \n`{3} matches exactly a newline followed by three three backticks (`).
    
    The regular expression matches a string that starts with three backticks followed by the string "python" and a newline character, then captures any sequence of characters (including whitespace and line breaks) until it encounters another newline and three backticks.
    """
    return re.findall("`{3}python\n([\s\S]*?)\n`{3}", contents)


def run(snippet: str, filename: str) -> SnippetExec:
    """
    Save the code snippets into a file and run through it.
    Return
    - `SnippetExec` which tells us:
        - Whether we are able to successfully run through the entire snippet
        - Any errors
        - Any output
    """
    success = True
    error = ""
    output = ""

    try:
        # Create a temporary file with every code block from the Markdown file.
        with open(filename, "w") as f:
            f.write(snippet)

        # Execute the code in the file and capture the output.
        result = subprocess.run(
            ["python3", filename],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=True,
            text=True,
        )

        # Store the output from execution.
        output = result.stdout
    except subprocess.CalledProcessError as e:
        success = False
        output = e.stdout
        error = e.stderr
    # Delete the temporary file
    os.remove(filename)
    return SnippetExec(snippet, success, output, error)


def should_skip_section(item):
    # Check to see if the section is excluded via `blacklist_sections`.
    for blacklist_section in blacklist_sections:
        current_section = item[0].split("gitbook/")[-1]
        if current_section == blacklist_section or current_section.startswith(
            blacklist_section + "/"
        ):
            print(">> SKIPPING ", current_section)
            return True


def should_skip_file(item):
    # Check to see if the section is excluded via `blacklist_files`.
    for blacklist_file in blacklist_files:
        if item == blacklist_file:
            print(">> SKIPPING ", item)
            return True


def remove_skipped_snippets(snippets):
    # Remove any snippet that shows up in `blacklist_snippets`.
    if file_name in blacklist_snippets.keys():
        return [snippet for snippet in snippets if snippet not in blacklist_snippets[file_name]]
    else:
        return snippets


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--docs-folder",
        dest="docs_folder",
        default="gitbook",
        action="store",
        help="Path to gitbook directory.",
    )

    parser.add_argument(
        "--run-subset",
        dest="run_subset",
        default=False,
        action="store_true",
        help="Run a subset of the gitbook sections.",
    )

    parser.add_argument(
        "--sections",
        dest="sections",
        default=[],
        nargs="+",
        action="store",
        help="Subset of gitbook sections to run. Expects a list of paths.",
    )

    args = parser.parse_args()

    # If we do not want to test every code snippet in the documentation, we can `run_subset` instead.
    if args.run_subset:
        # The subset of sections we run is specified in `sections_subset`. The for loop expects a list of tuples so we add a dummy 0.
        args.sections = [(section, 0) for section in args.sections_subset]
    else:
        # Otherwise, we look at all files in all sections in the `docs_folder`.
        args.sections = os.walk(args.docs_folder)

    successfully_ran_all = True
    for item in args.sections:
        if should_skip_section(item):
            continue

        for file in glob(os.path.join(item[0], "*.md")):
            file_name = file.split("gitbook/")[-1]

            if should_skip_file(file_name):
                continue

            snippets = remove_skipped_snippets(get_code(file))

            # If we still have code to run, run it.
            if len(snippets) > 0:
                temp_file_name = file.replace(r"/", "_")[:-2] + "py"
                snippet_result = run("\n".join(snippets), temp_file_name)

                if not snippet_result.success:
                    # Snippet failed to run. Display relevant information for debugging.
                    successfully_ran_all = False
                    print(">> FAILED   ", file_name)
                    # Display the code
                    print("-" * 25 + "[CODE]" + "-" * 25)
                    print(snippet_result.snippet)
                    # Display the output
                    print("-" * 25 + "[OUTPUT]" + "-" * 25)
                    print(snippet_result.output)
                    # Display the error
                    print("-" * 25 + "[ERROR]" + "-" * 25)
                    print(snippet_result.error)
                else:
                    print(">> OK       ", file_name)
            else:
                print(">> SKIPPING  (no Python snippets found)", file_name)

    print("Took", time.time() - start, "s.")
    if successfully_ran_all:
        print("Congrats! You've successfully ran all code snippets in the documentation.")
        exit(0)
    exit(1)
