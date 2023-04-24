"""
Installs aqueduct from the local repo. Run with `python3 scripts/install_local.py` from the root directory of the aqueduct repo.

Requirements:
- `aqueduct-ml` must already be installed.
- `aqueduct start` must have been run at least once since the last `aqueduct clear`. (~/.aqueduct must exist)
- The aqueduct server must not be running.

After this script completes, running `aqueduct start` will start with the local changes.

If you don't specify any component flag, the script will update all components. Keep in mind that UI
takes longer to update.
"""

import argparse
import os
import re
import shutil
import subprocess
import sys
from os import listdir
from os.path import isdir, isfile, join

base_directory = join(os.environ["HOME"], ".aqueduct")
server_directory = join(os.environ["HOME"], ".aqueduct", "server")
ui_directory = join(os.environ["HOME"], ".aqueduct", "ui")

# Make sure to update this if there is any schema change we want to include in the upgrade.
SCHEMA_VERSION = "26"


def execute_command(args, cwd=None):
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "-u",
        "--ui",
        dest="update_ui",
        default=False,
        action="store_true",
        help="Whether to build and replace UI files.",
    )

    parser.add_argument(
        "-g",
        "--gobinary",
        dest="update_go_binary",
        default=False,
        action="store_true",
        help="Whether to build and replace Golang binaries.",
    )

    parser.add_argument(
        "-s",
        "--sdk",
        dest="update_sdk",
        default=False,
        action="store_true",
        help="Whether to build and replace the Python SDK.",
    )

    parser.add_argument(
        "-e",
        "--executor",
        dest="update_executor",
        default=False,
        action="store_true",
        help="Whether to build and replace the Python executor.",
    )

    args = parser.parse_args()

    if not (args.update_ui or args.update_go_binary or args.update_sdk or args.update_executor):
        args.update_ui = True
        args.update_go_binary = True
        args.update_sdk = True
        args.update_executor = True

    cwd = os.getcwd()
    if not cwd.endswith("aqueduct"):
        print("Current directory should be the root directory of the aqueduct repo.")
        print("Your working directory is %s" % cwd)
        exit(1)

    if not isdir(base_directory):
        print("~/.aqueduct must exist.")
        exit(1)

    # TODO(kenxu): Can be removed once all development environments have this folder already.
    preview_outputs_directory = os.path.join(server_directory, "storage", "preview")
    if not os.path.isdir(preview_outputs_directory):
        os.mkdir(preview_outputs_directory)

    # Force the env to be "dev", so that we don't have to manually set the env when starting the server.
    env_file_path = os.path.join(server_directory, "config", "env")
    with open(env_file_path, "w") as f:
        f.write("dev")

    # Install the local SDK.
    if args.update_sdk:
        print("Updating the Python SDK...")
        prev_pwd = os.environ["PWD"]
        os.environ["PWD"] = join(os.environ["PWD"], "sdk")
        execute_command(["python3", "-m", "pip", "install", "."], cwd=join(cwd, "sdk"))
        os.environ["PWD"] = prev_pwd

    # Install the local python operators.
    if args.update_executor:
        print("Updating the Python executor...")
        prev_pwd = os.environ["PWD"]
        os.environ["PWD"] = join(os.environ["PWD"], "src/python")
        execute_command(["python3", "-m", "pip", "install", "."], cwd=join(cwd, "src", "python"))
        os.environ["PWD"] = prev_pwd

        execute_command(
            [
                "cp",
                "./src/python/aqueduct_executor/start-function-executor.sh",
                join(server_directory, "bin"),
            ]
        )

        execute_command(
            [
                "cp",
                "./src/python/aqueduct_executor/operators/airflow/dag.template",
                join(server_directory, "bin"),
            ]
        )

    # Build and replace backend binaries.
    if args.update_go_binary:
        print("Updating Golang binaries...")
        execute_command(["make", "server"], cwd=join(cwd, "src"))
        execute_command(["make", "executor"], cwd=join(cwd, "src"))
        execute_command(["make", "migrator"], cwd=join(cwd, "src"))
        if isfile(join(server_directory, "bin/server")):
            execute_command(["rm", join(server_directory, "bin/server")])
        if isfile(join(server_directory, "bin/executor")):
            execute_command(["rm", join(server_directory, "bin/executor")])
        if isfile(join(server_directory, "bin/migrator")):
            execute_command(["rm", join(server_directory, "bin/migrator")])

        execute_command(["cp", "./src/build/server", join(server_directory, "bin/server")])
        execute_command(["cp", "./src/build/executor", join(server_directory, "bin/executor")])
        execute_command(["cp", "./src/build/migrator", join(server_directory, "bin/migrator")])

        # Run the migrator to update to the latest schema
        execute_command(
            [join(server_directory, "bin/migrator"), "--type", "sqlite", "goto", SCHEMA_VERSION]
        )

    # Build and replace UI files.
    if args.update_ui:
        UI_PATH = "src/ui"
        UI_COMMON_PATH = UI_PATH + "/common"
        UI_APP_PATH = UI_PATH + "/app"

        print("Updating UI files...")
        execute_command(["rm", "-rf", "node_modules"], cwd=join(cwd, UI_COMMON_PATH))
        execute_command(["rm", "-rf", ".parcel-cache"], cwd=join(cwd, UI_COMMON_PATH))
        execute_command(["rm", "-rf", "dist"], cwd=join(cwd, UI_COMMON_PATH))
        execute_command(["npm", "install", "--force"], cwd=join(cwd, UI_COMMON_PATH))
        execute_command(["npm", "link"], cwd=join(cwd, UI_COMMON_PATH))
        execute_command(["rm", "-rf", "node_modules"], cwd=join(cwd, UI_APP_PATH))
        execute_command(["rm", "-rf", ".parcel-cache"], cwd=join(cwd, UI_APP_PATH))
        execute_command(["rm", "-rf", "dist"], cwd=join(cwd, UI_APP_PATH))
        execute_command(["npm", "link", "@aqueducthq/common"], cwd=join(cwd, UI_APP_PATH))
        execute_command(["make", "dist"], cwd=join(cwd, "src/ui"))

        files = [f for f in listdir(ui_directory) if isfile(join(ui_directory, f))]
        for f in files:
            if not f == "__version__":
                execute_command(["rm", f], cwd=ui_directory)

        shutil.copytree(
            join(cwd, "src", "ui", "app", "dist", "default"), ui_directory, dirs_exist_ok=True
        )

        # To prevent unnecessary files from getting into our releases
        # Will replace the react-code-block component soon (next week) to avoid this concern completely
        files = [f for f in listdir(ui_directory) if isfile(join(ui_directory, f))]
        fileNameRegex = re.compile(
            r"^(python|core|markup|clike|javascript|css|index|favicon)\..*(html|js|css|map|ico)$"
        )
        for f in files:
            if not fileNameRegex.search(f) and not f == "__version__":
                execute_command(["rm", f], cwd=ui_directory)

    print("Successfully installed aqueduct from local repo!")
