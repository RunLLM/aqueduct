import os
import re
import argparse
import subprocess
from glob import glob
from pathlib import Path
from typing import NamedTuple, List

class SnippetExec(NamedTuple):
    # The result from executing the code snippets sequentially.
    # Returned from `run` and stores useful information for debugging that is displayed if a file fails to run to the end successfully.
    
    # All the Python code block in the Markdown file joined in order of appearance.
    snippet: str
    # Whether or not every snippet is successfully executed.
    success: bool
    # The index of the last successfully ran block.
    last_successful_block: int
    # The error message.
    error: str

# Not in-scope for testing now
blacklist_sections = [
    "integrations",  # Requires setting up the integrations which currently cannot be done through the SDK alone
    "installation-and-configuration/installing-aqueduct",  # Specific examples that relate to screenshots of the external setups (e.g. using the IP address displayed on the screenshot)
    "example-workflows",  # Already tested elsewhere
]

blacklist_files = [
    "operators/configuring-resource-constraints.md",  # Header snippets. The GPU Access one require the operator to be executed on Kubernetes
    "workflows/deleting-a-workflow.md",  # Referencing workflow UUIDs.
    "operators/specifying-a-requirements.txt.md", # Requires specific requirements.txt
]

blacklist_snippets = {
    "workflows/managing-workflow-schedules.md": [
# Skipping because references workflows by UUID.
"""workflow_a = client.flow('8fb25dc4-62ed-44a3-872d-c3ff988c8dd3')

# source_flow can be a Flow object, workflow name, or workflow ID
flow = client.publish_flow(name='workflow_b', 
                           artifacts=[data],
                           source_flow=source_flow)""",
# Skipping because references workflows by UUID.
"""workflow_id = "0c007eff-6ae0-4a1a-a114-5f16164ffcdf" # Set your workflow ID here.
client.trigger(id=workflow_id)""",
    ],
}
def get_code(page: str) -> List[str]:
    '''
    Use an regular expression to find all Python code blocks in the page.
    Return
    - List of each code block found in the page in order of appearance
    '''
    contents = Path(page).read_text()

    # The regular expression is used to extract code blocks written in the Python language that are enclosed in a triple backtick (```) fence.

    # - `{3} matches exactly three occurrences of the previous character or group, which in this case is the backtick (`).
    # - python\n matches the string "python" followed by a newline character.
    # - ([\s\S]*?) is a capture group that matches any sequence of characters, including whitespace characters and line breaks. The *? quantifier means "match zero or more of the preceding token, but as few as possible".
    # - \n`{3} matches exactly a newline followed by three three backticks (`).
    
    # The regular expression matches a string that starts with three backticks followed by the string "python" and a newline character, then captures any sequence of characters (including whitespace and line breaks) until it encounters another newline and three backticks.
    
    return re.findall("`{3}python\n([\s\S]*?)\n`{3}", contents)

def run(snippet: str, filename: str) -> SnippetExec:
    '''
    Save the code snippets into a file and run through it.
    Return
    - `SnippetExec` which tells us:
        - Whether we are able to successfully run through the entire snippet
        - Any errors
        - Any output
    '''
    success = True
    error = ""
    output = ""

    try:
        # Create a temporary file with every code block from the Markdown file.
        with open(filename, "w") as f:
            f.write(snippet)

        # Execute the code in the file and capture the output.
        result = subprocess.run(["python3", filename], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True, text=True)

        # Store the output from execution.
        output = result.stdout
    except subprocess.CalledProcessError as e:
        success = False
        output = e.stdout
        error = e.stderr
    # Delete the temporary file
    os.remove(filename)
    return SnippetExec(snippet, success, output, error)

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
        nargs='+',
        action="store",
        help="Subset of gitbook sections to run.",
    )

    args = parser.parse_args()

    if args.run_subset:
        args.sections = [(section, 0) for section in args.sections_subset]
    else:
        args.sections = os.walk(args.docs_folder)

    snippets_by_page = {}

    successfully_ran_all = True
    for item in args.sections:
        skip = False
        for blacklist_section in blacklist_sections:
            current_section = item[0].split("gitbook/")[-1]
            if current_section == blacklist_section or current_section.startswith(blacklist_section+"/"):
                skip = True
                print(">> SKIPPING ", current_section)
                break
        if not skip:
            for file in glob(os.path.join(item[0], '*.md')):
                file_name = file.split("gitbook/")[-1]
                skip = False
                for blacklist_file in blacklist_files:
                    if file_name == blacklist_file:
                        skip = True
                        print(">> SKIPPING ", file_name)
                if not skip:
                    snippets = get_code(file)
                    if file_name in blacklist_snippets.keys():
                        snippets = [snippet for snippet in snippets if snippet not in blacklist_snippets[file_name]]
                    if len(snippets) > 0:
                        snippets_by_page[file] = run("\n".join(snippets), file.replace(r"/", "_")[:-2] + "py")
                        if not snippets_by_page[file].success:
                            # Snippet failed to run. Display relevant information for debugging. 
                            successfully_ran_all = False
                            print(">> FAILED   ", file_name)
                            # Display the code
                            print("-"*25 + "[CODE]" + "-"*25)
                            print(snippets_by_page[file].snippet)
                            # Display the output
                            print("-"*25 + "[OUTPUT]" + "-"*25)
                            print(snippets_by_page[file].output)
                            # Display the error
                            print("-"*25 + "[ERROR]" + "-"*25)
                            print(snippets_by_page[file].error)
                        else:
                            print(">> OK       ", file_name)
                    else:
                        print(">> SKIPPING  (no Python snippets found)", file_name)


    if successfully_ran_all:
        print("Congrats! You've successfully ran all code snippets in the documentation.")