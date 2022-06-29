from setuptools import find_packages, setup
from setuptools.command.develop import develop
from setuptools.command.install import install
from setuptools.command.egg_info import egg_info
from pathlib import Path
import os
import string
import random
import subprocess
import sys
import platform

base_directory = os.path.join(os.environ["HOME"], ".aqueduct")
server_directory = os.path.join(os.environ["HOME"], ".aqueduct", "server")
ui_directory = os.path.join(os.environ["HOME"], ".aqueduct", "ui")

package_version = "0.0.3"

def update_config_yaml(file):
    s=string.ascii_uppercase+string.digits
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
    print("Updated configurations.")

def execute_command(args, cwd=None):
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)

def generate_version_file(file_path):
    with open(file_path, 'w') as f:
        f.write(package_version)
    print("Wrote version to file.")

# Returns a bool indicating whether we need to perform a version upgrade.
def require_update(file_path):
    if not os.path.isfile(file_path):
        return True
    with open(file_path, 'r') as f:
        current_version = f.read()
        if package_version < current_version:
            raise Exception("Attempting to install an older version %s but found existing newer version %s" % (package_version, current_version))
        elif package_version == current_version:
            return False
        else:
            return True

def update_executable_permissions():
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "server")])
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "executor")])
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "migrator")])

def download_server_binaries():
    print("Downloading server binaries.")
    s3_prefix = "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/%s/server" % package_version
    execute_command(["rm", "-rf", os.path.join(server_directory, "bin")])
    os.mkdir(os.path.join(server_directory, "bin"))

    system = platform.system()
    arch = platform.machine()
    if system == "Linux" and arch == "x86_64":
        print("Operating system is Linux with architecture amd64.")
        execute_command(["curl", os.path.join(s3_prefix, "bin/linux_amd64/server"), "--output", os.path.join(server_directory, "bin/server")])
        execute_command(["curl", os.path.join(s3_prefix, "bin/linux_amd64/executor"), "--output", os.path.join(server_directory, "bin/executor")])
        execute_command(["curl", os.path.join(s3_prefix, "bin/linux_amd64/migrator"), "--output", os.path.join(server_directory, "bin/migrator")])
    elif system == "Darwin" and arch == "x86_64":
        print("Operating system is Mac with architecture amd64.")
        execute_command(["curl", os.path.join(s3_prefix, "bin/darwin_amd64/server"), "--output", os.path.join(server_directory, "bin/server")])
        execute_command(["curl", os.path.join(s3_prefix, "bin/darwin_amd64/executor"), "--output", os.path.join(server_directory, "bin/executor")])
        execute_command(["curl", os.path.join(s3_prefix, "bin/darwin_amd64/migrator"), "--output", os.path.join(server_directory, "bin/migrator")])
    elif system == "Darwin" and arch == "arm64":
        print("Operating system is Mac with architecture arm64.")
        execute_command(["curl", os.path.join(s3_prefix, "bin/darwin_arm64/server"), "--output", os.path.join(server_directory, "bin/server")])
        execute_command(["curl", os.path.join(s3_prefix, "bin/darwin_arm64/executor"), "--output", os.path.join(server_directory, "bin/executor")])
        execute_command(["curl", os.path.join(s3_prefix, "bin/darwin_arm64/migrator"), "--output", os.path.join(server_directory, "bin/migrator")])
    else:
        raise Exception("Unsupported operating system and architecture combination: %s, %s" % (system, arch))
    
    execute_command(["curl", os.path.join(s3_prefix, "bin/start-function-executor.sh"), "--output", os.path.join(server_directory, "bin/start-function-executor.sh")])
    execute_command(["curl", os.path.join(s3_prefix, "bin/install_sqlserver_ubuntu.sh"), "--output", os.path.join(server_directory, "bin/install_sqlserver_ubuntu.sh")])

def update_ui_version():
    print("Updating UI version to %s" % package_version)
    try:
        execute_command(["rm", "-rf", ui_directory])
        os.mkdir(ui_directory)
        generate_version_file(os.path.join(ui_directory, "__version__"))
        s3_prefix = "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/%s/ui" % package_version
        execute_command(["curl", os.path.join(s3_prefix, "ui.zip"), "--output", os.path.join(ui_directory, "ui.zip")])
        execute_command(["unzip", os.path.join(ui_directory, "ui.zip"), "-d", ui_directory])
        execute_command(["rm", os.path.join(ui_directory, "ui.zip")])
    except Exception as e:
        print(e)
        execute_command(["rm", "-rf", ui_directory])
        exit(1)

def update_server_version():
    print("Updating server version to %s" % package_version)
    if os.path.isfile(os.path.join(server_directory, "__version__")):
        execute_command(["rm", os.path.join(server_directory, "__version__")])
    generate_version_file(os.path.join(server_directory, "__version__"))

    download_server_binaries()
    update_executable_permissions()

    execute_command([os.path.join(server_directory, "bin", "migrator"), "--type", "sqlite", "goto", "9"])

def updates():
    if not os.path.isdir(base_directory):
        os.mkdir(base_directory)

    if not os.path.isdir(ui_directory) or require_update(os.path.join(ui_directory, "__version__")):
        update_ui_version()
    
    if not os.path.isdir(server_directory):
        try:
            directories = [
                server_directory,
                os.path.join(server_directory, "db"),
                os.path.join(server_directory, "storage"),
                os.path.join(server_directory, "storage", "operators"),
                os.path.join(server_directory, "vault"),
                os.path.join(server_directory, "bin"),
                os.path.join(server_directory, "config"),
                os.path.join(server_directory, "logs"),
            ]

            for directory in directories:
                os.mkdir(directory)

            update_server_version()

            s3_prefix = "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/%s/server" % package_version
            execute_command(["curl", os.path.join(s3_prefix, "config/config.yml"), "--output", os.path.join(server_directory, "config/config.yml")])
            update_config_yaml(os.path.join(server_directory, "config", "config.yml"))
            execute_command(["curl", os.path.join(s3_prefix, "db/demo.db"), "--output", os.path.join(server_directory, "db/demo.db")])

            print("Finished initializing Aqueduct base directory.")
        except Exception as e:
            print(e)
            execute_command(["rm", "-rf", server_directory])
            exit(1)

    if require_update(os.path.join(server_directory, "__version__")):
        try:
            update_server_version()
        except Exception as e:
            print(e)
            if os.path.isfile(os.path.join(server_directory, "__version__")):
                execute_command(["rm", os.path.join(server_directory, "__version__")])
            exit(1)

class DevelopCommand(develop):
    """Post-installation for development mode."""
    def run(self):
        develop.run(self)
        updates()

class InstallCommand(install):
    """Post-installation for installation mode."""
    def run(self):
        install.run(self)
        updates()

class EggInfoCommand(egg_info):
    """Post-installation for egg info mode."""
    def run(self):
        egg_info.run(self)
        updates()

install_requires = open("requirements.txt").read().strip().split("\n")

readme_path = Path(os.environ["PWD"], "../../README.md")
long_description = open(readme_path).read()

setup(
    name="aqueduct-ml",
    version="0.0.3",
    install_requires=install_requires,
    scripts=["bin/aqueduct"],
    packages=find_packages(),
    description="Prediction Infrastructure for Data Scientists",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://www.aqueducthq.com/",
    license="Apache License 2.0",
    author="Aqueduct, Inc.",
    author_email="hello@aqueducthq.com",
    classifiers=[
        "Programming Language :: Python :: 3",
    ],
    cmdclass={
        'develop': DevelopCommand,
        'install': InstallCommand,
        'egg_info': EggInfoCommand,
    },
    python_requires=">=3.7",
)
