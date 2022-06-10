"""
Installs aqueduct from the local repo. Run with `python3 scripts/install_local.py`.

Requirements:
- `aqueduct-ml` must already be installed.
- `aqueduct start` must have been run at least once since the last `aqueduct clear`. (~/.aqueduct/server must exist)
- The aqueduct server must not be running.

After this script completes, running `aqueduct start` will start the backend with the local changes.
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

    if not os.path.isdir(server_directory):
        print("~/.aqueduct/server must exist.")
        exit(1)

    # Build the local backend binaries.
    execute_command(["make", "server"], cwd=os.path.join(cwd, "src"))
    execute_command(["make", "executor"], cwd=os.path.join(cwd, "src"))
    execute_command(["make", "migrator"], cwd=os.path.join(cwd, "src"))
    if os.path.isfile(os.path.join(server_directory, "bin/server")):
        execute_command(["rm", os.path.join(server_directory, "bin/server")])
    if os.path.isfile(os.path.join(server_directory, "bin/executor")):
        execute_command(["rm", os.path.join(server_directory, "bin/executor")])
    if os.path.isfile(os.path.join(server_directory, "bin/migrator")):
        execute_command(["rm", os.path.join(server_directory, "bin/migrator")])

    execute_command(["cp", "./src/build/server", os.path.join(server_directory, "bin/server")])
    execute_command(["cp", "./src/build/executor", os.path.join(server_directory, "bin/executor")])

    # Install the local SDK.
    os.environ["PWD"] = os.path.join(os.environ["PWD"], "sdk")
    execute_command(["pip", "install", "."], cwd=os.path.join(cwd, "sdk"))

    # Install the local python operators.
    os.environ["PWD"] = os.path.join(os.environ["PWD"], "../src/python")
    execute_command(["pip", "install", "."], cwd=os.path.join(cwd, "src", "python"))

    print("Successfully installed aqueduct from local repo!")
