import os
import re
import argparse
import subprocess
from glob import glob
from pathlib import Path
from typing import NamedTuple, List

class SnippetExec(NamedTuple):
    snippet: List[str]
    success: bool
    output: str
    last_successful_block: int
    error: str

# Not in-scope for testing now
blacklist_sections = [
    "integrations",
    "api-reference",
    "installation-and-configuration",
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
def get_code(page):
    contents = Path(page).read_text()
    return re.findall("`{3}python\n([\s\S]*?)\n`{3}", contents)

def run_in_order(snippet_list, filename):
    success = True
    error = ""
    output = ""
    i = len(snippet_list)

    try:
        # Create a temporary file
        with open(filename, "w") as f:
            f.write("\n".join(snippet_list))

        # Execute the code in the file and capture the output
        result = subprocess.run(["python3", filename], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True, text=True)

        output = result.stdout
    except subprocess.CalledProcessError as e:
        if i <= 1:
            success = False
            output = e.stdout
            error = e.stderr
        else:
            # Determine which block is the issue.
            for i in range(1, i):
                try:
                    with open(filename, "w") as f:
                        f.write("\n".join(snippet_list[:i]))
                    result = subprocess.run(["python3", filename], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True, text=True)
                except subprocess.CalledProcessError as e:
                    success = False
                    output = e.stdout
                    error = e.stderr
                    i -= 1
                    break
    # Delete the temporary file
    os.remove(filename)
    return SnippetExec(snippet_list, success, output, i, error)

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
                        snippets_by_page[file] = run_in_order(snippets, file.replace(r"/", "_")[:-2] + "py")
                        if not snippets_by_page[file].success:
                            successfully_ran_all = False
                            print(">> FAILED   ", file_name)
                            print("-"*25 + "[CODE]" + "-"*25)
                            for i, snippet in enumerate(snippets_by_page[file].snippet):
                                if i == snippets_by_page[file].last_successful_block:
                                    print("-"*25 + "[FAILED BLOCK]" + "-"*25)
                                print(snippet)
                                if i == snippets_by_page[file].last_successful_block:
                                    break
                            print("-"*25 + "[ERROR]" + "-"*25)
                            print(snippets_by_page[file].error)
                        else:
                            print(">> OK       ", file_name)
                    else:
                        print(">> SKIPPING  (no Python snippets found)", file_name)


    if successfully_ran_all:
        print("Congrats! You've successfully ran all code snippets in the documentation.")