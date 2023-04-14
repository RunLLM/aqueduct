import json
import os
import subprocess
import sys

CELL_CODE_HEADER_TEMPLATE = 'print("Cell %d")\n'
# These are the prefixes that we use to identify and extract client credentials from the notebook.
SERVER_ADDRESS_CODE_SNIPPET = "address = "
API_KEY_SNIPPET = "aqueduct.get_apikey()"
ABBR_API_KEY_SNIPPET = "aq.get_apikey()"


# Pull out the client credential value in the notebook, formatted like "<credential_prefix> <value>\n".
# Strips out any quotes.
def extract_credential(code: str, credential_prefix: str) -> str:
    start_idx = code.find(credential_prefix)
    if start_idx < 0:
        return ""
    end_idx = code.find("\n", start_idx)
    return code[start_idx + len(credential_prefix) : end_idx].strip('"')


def replace_server_addr(code: str, addr: str) -> str:
    old_server_address = extract_credential(code, SERVER_ADDRESS_CODE_SNIPPET)
    if old_server_address:
        code = code.replace(old_server_address, addr)
    return code


def replace_api_key(code: str, api_key: str) -> str:
    code = code.replace(API_KEY_SNIPPET, f'"{api_key}"')
    code = code.replace(ABBR_API_KEY_SNIPPET, f'"{api_key}"')
    return code


def deploy(dir: str, name: str, tmp_name: str, addr: str, api_key: str) -> None:
    current_dir = os.getcwd()
    os.chdir(dir)
    with open(name, "r") as f:
        notebook = json.load(f)

    # Pull out the notebook code.
    code_blocks = [c["source"] for c in notebook["cells"] if c["cell_type"] == "code"]
    code_block_list = [
        "".join([CELL_CODE_HEADER_TEMPLATE % i] + block) for i, block in enumerate(code_blocks)
    ]
    code = "\n\n\n".join(code_block_list)
    code = replace_server_addr(code, addr)
    code = replace_api_key(code, api_key)
    with open(tmp_name, "w") as f:
        f.write(code)

    process = subprocess.run([sys.executable, tmp_name])

    os.remove(tmp_name)
    os.chdir(current_dir)
    if process.returncode:
        raise Exception(f"Error executing notebook {name}")