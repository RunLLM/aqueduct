"""
Installs aqueduct from the local repo. Run with `python3 scripts/install_local.py`.

Requirements:
- `aqueduct-ml` must already be installed.
- `aqueduct server` must have been run at least once since the last `aqueduct clear`. (~/.aqueduct/server must exist)
- The aqueduct server must not be running.

After this script completes, running `aqueduct server` will start the backend with the local changes.
The sdk will also be updated with any local changes.
"""

import os
import subprocess
import sys

base_directory = os.path.join(os.environ["HOME"], ".aqueduct")
server_directory = os.path.join(os.environ["HOME"], ".aqueduct", "server")


def execute_command(args, cwd=None):
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


if __name__ == "__main__":
    print("Current directory should be the root directory of the aqueduct repo.")
    cwd = os.getcwd()
    if not cwd.endswith("aqueduct"):
        print("Your working directory is %s" % cwd)
        exit(1)

    # Create the ~/.aqueduct directory is it does not already exist. Copied from `/src/python/bin/aqueduct`.
    if not os.path.isdir(server_directory):
        print("~/.aqueduct/server must exist.")
        exit(1)

    # Build the local backend binaries.
    execute_command(["make", "server"], cwd=os.path.join(cwd, "src"))
    execute_command(["cp", "./src/build/server", os.path.join(server_directory, "bin", "server")])
    execute_command(["make", "executor"], cwd=os.path.join(cwd, "src"))
    execute_command(["cp", "./src/build/executor", os.path.join(server_directory, "bin", "executor")])

    # Install the local python operators.
    execute_command(["pip", "install", ".", "--user"], cwd=os.path.join(cwd, "src", "python"))

    # Install the local SDK.
    execute_command(["pip", "install", ".", "--user"], cwd=os.path.join(cwd, "sdk"))

    print("Successfully installed aqueduct from local repo!")
