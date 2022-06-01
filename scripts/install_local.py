"""
Installs aqueduct from the local repo. Run with `python3 scripts/install_loca.py`.

Requires no aqueduct server to be running.

After this script completes, running `aqueduct server` will start the backend with the local changes.
The sdk will also be updated with any local changes.
"""

import os
import platform
import random
import string
import subprocess
import sys
import traceback as tb

base_directory = os.path.join(os.environ["HOME"], ".aqueduct")
server_directory = os.path.join(os.environ["HOME"], ".aqueduct", "server")


def execute_command(args, cwd=None):
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


def update_executable_permissions():
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "server")])
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "executor")])
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "migrator")])


def update_config_yaml(file):
    s = string.ascii_uppercase+string.digits
    encryption_key = ''.join(random.sample(s,32))
    api_key = ''.join(random.sample(s,32))

    with open(file, "r") as sources:
        lines = sources.readlines()
    with open(file, "w") as sources:
        for line in lines:
            if "<BASE_PATH>" in line:
                sources.write(line.replace("<BASE_PATH>", server_directory))
            elif "<ENCRYPTION_KEY>" in line:
                sources.write(line.replace("<ENCRYPTION_KEY>", encryption_key))
            elif "<API_KEY>" in line:
                sources.write(line.replace("<API_KEY>", api_key))
            else:
                sources.write(line)


if __name__ == "__main__":
    print("Current directory should be the root directory of the aqueduct repo.")
    cwd = os.getcwd()
    if not cwd.endswith("aqueduct"):
        print("Your working directory is %s" % cwd)
        exit(1)

    # Create the ~/.aqueduct directory is it does not already exist. Copied from `/src/python/bin/aqueduct`.
    if not os.path.isdir(server_directory):
        try:
            directories = [
                base_directory,
                server_directory,
                os.path.join(server_directory, "db"),
                os.path.join(server_directory, "storage"),
                os.path.join(server_directory, "storage", "operators"),
                os.path.join(server_directory, "vault"),
                os.path.join(server_directory, "bin"),
                os.path.join(server_directory, "config"),
            ]

            for directory in directories:
                if not os.path.isdir(directory):
                    os.mkdir(directory)

            system = platform.system()
            arch = platform.machine()
            if system == "Linux" and arch == "x86_64":
                print("Operating system is Linux with architecture amd64.")
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/linux_amd64/server", "--output", os.path.join(server_directory, "bin", "server")])
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/linux_amd64/executor", "--output", os.path.join(server_directory, "bin", "executor")])
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/linux_amd64/migrator", "--output", os.path.join(server_directory, "bin", "migrator")])
            elif system == "Darwin" and arch == "x86_64":
                print("Operating system is Mac with architecture amd64.")
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/darwin_amd64/server", "--output", os.path.join(server_directory, "bin", "server")])
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/darwin_amd64/executor", "--output", os.path.join(server_directory, "bin", "executor")])
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/darwin_amd64/migrator", "--output", os.path.join(server_directory, "bin", "migrator")])
            elif system == "Darwin" and arch == "arm64":
                print("Operating system is Mac with architecture arm64.")
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/darwin_arm64/server", "--output", os.path.join(server_directory, "bin", "server")])
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/darwin_arm64/executor", "--output", os.path.join(server_directory, "bin", "executor")])
                execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/darwin_arm64/migrator", "--output", os.path.join(server_directory, "bin", "migrator")])
            else:
                raise Exception("Unsupported operating system and architecture combination: %s, %s" % (system, arch))

            update_executable_permissions()

            execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/start-function-executor.sh", "--output", os.path.join(server_directory, "bin", "start-function-executor.sh")])
            execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/bin/install_sqlserver_ubuntu.sh", "--output", os.path.join(server_directory, "bin", "install_sqlserver_ubuntu.sh")])
            execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/db/demo.db", "--output", os.path.join(server_directory, "db", "demo.db")])
            execute_command(["curl", "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/config/config.yml", "--output", os.path.join(server_directory, "config", "config.yml")])

            update_config_yaml(os.path.join(server_directory, "config", "config.yml"))

            execute_command([os.path.join(server_directory, "bin", "migrator"), "--type", "sqlite", "goto", "8"])

            print("Finished initializing Aqueduct base directory.")
        except Exception as e:
            print(e)
            tb.print_tb(e.__traceback__)
            execute_command(["rm", "-rf", server_directory])
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
