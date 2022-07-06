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
import shutil
import requests
import zipfile

base_directory = os.path.join(os.environ["HOME"], ".aqueduct")
server_directory = os.path.join(os.environ["HOME"], ".aqueduct", "server")
ui_directory = os.path.join(os.environ["HOME"], ".aqueduct", "ui")

package_version = "0.0.3"
s3_server_prefix = (
    "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/%s/server" % package_version
)
s3_ui_prefix = "https://aqueduct-ai.s3.us-east-2.amazonaws.com/assets/%s/ui" % package_version


def update_config_yaml(file):
    s = string.ascii_uppercase + string.digits
    encryption_key = "".join(random.sample(s, 32))
    api_key = "".join(random.sample(s, 32))

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
    with open(file_path, "w") as f:
        f.write(package_version)
    print("Wrote version to file.")


# Returns a bool indicating whether we need to perform a version upgrade.
def require_update(file_path):
    if not os.path.isfile(file_path):
        return True
    with open(file_path, "r") as f:
        current_version = f.read()
        if package_version < current_version:
            raise Exception(
                "Attempting to install an older version %s but found existing newer version %s"
                % (package_version, current_version)
            )
        elif package_version == current_version:
            return False
        else:
            return True


def update_executable_permissions():
    os.chmod(os.path.join(server_directory, "bin", "server"), 0o755)
    os.chmod(os.path.join(server_directory, "bin", "executor"), 0o755)
    os.chmod(os.path.join(server_directory, "bin", "migrator"), 0o755)


def download_server_binaries(architecture):
    with open(os.path.join(server_directory, "bin/server"), "wb") as f:
        f.write(requests.get(os.path.join(s3_server_prefix, f"bin/{architecture}/server")).content)
    with open(os.path.join(server_directory, "bin/executor"), "wb") as f:
        f.write(
            requests.get(os.path.join(s3_server_prefix, f"bin/{architecture}/executor")).content
        )
    with open(os.path.join(server_directory, "bin/migrator"), "wb") as f:
        f.write(
            requests.get(os.path.join(s3_server_prefix, f"bin/{architecture}/migrator")).content
        )
    with open(os.path.join(server_directory, "bin/start-function-executor.sh"), "wb") as f:
        f.write(
            requests.get(os.path.join(s3_server_prefix, "bin/start-function-executor.sh")).content
        )
    with open(os.path.join(server_directory, "bin/install_sqlserver_ubuntu.sh"), "wb") as f:
        f.write(
            requests.get(os.path.join(s3_server_prefix, "bin/install_sqlserver_ubuntu.sh")).content
        )


def setup_server_binaries():
    print("Downloading server binaries.")
    server_bin_path = os.path.join(server_directory, "bin")
    shutil.rmtree(server_bin_path, ignore_errors=True)
    os.mkdir(server_bin_path)

    system = platform.system()
    arch = platform.machine()
    if system == "Linux" and arch == "x86_64":
        print("Operating system is Linux with architecture amd64.")
        download_server_binaries("linux_amd64")
    elif system == "Darwin" and arch == "x86_64":
        print("Operating system is Mac with architecture amd64.")
        download_server_binaries("darwin_amd64")
    elif system == "Darwin" and arch == "arm64":
        print("Operating system is Mac with architecture arm64.")
        download_server_binaries("darwin_arm64")
    else:
        raise Exception(
            "Unsupported operating system and architecture combination: %s, %s" % (system, arch)
        )


def update_ui_version():
    print("Updating UI version to %s" % package_version)
    try:
        shutil.rmtree(ui_directory, ignore_errors=True)
        os.mkdir(ui_directory)
        generate_version_file(os.path.join(ui_directory, "__version__"))
        ui_zip_path = os.path.join(ui_directory, "ui.zip")
        with open(ui_zip_path, "wb") as f:
            # We detect whether the server is running on a SageMaker instance by checking if the
            # directory /home/ec2-user/SageMaker exists. This is hacky but we couldn't find a better
            # solution at the moment.
            if os.path.isdir(os.path.join(os.sep, "home", "ec2-user", "SageMaker")):
                f.write(requests.get(os.path.join(s3_ui_prefix, "sagemaker", "ui.zip")).content)
            else:
                f.write(requests.get(os.path.join(s3_ui_prefix, "default", "ui.zip")).content)
        with zipfile.ZipFile(ui_zip_path, "r") as zip:
            zip.extractall(ui_directory)
        os.remove(ui_zip_path)
    except Exception as e:
        print(e)
        shutil.rmtree(ui_directory, ignore_errors=True)
        exit(1)


def update_server_version():
    print("Updating server version to %s" % package_version)

    version_file = os.path.join(server_directory, "__version__")
    if os.path.isfile(version_file):
        os.remove(version_file)
    generate_version_file(version_file)

    setup_server_binaries()
    update_executable_permissions()

    execute_command(
        [os.path.join(server_directory, "bin", "migrator"), "--type", "sqlite", "goto", "9"]
    )


def update():
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

            with open(os.path.join(server_directory, "config/config.yml"), "wb") as f:
                f.write(requests.get(os.path.join(s3_server_prefix, "config/config.yml")).content)

            update_config_yaml(os.path.join(server_directory, "config", "config.yml"))

            with open(os.path.join(server_directory, "db/demo.db"), "wb") as f:
                f.write(requests.get(os.path.join(s3_server_prefix, "db/demo.db")).content)

            print("Finished initializing Aqueduct base directory.")
        except Exception as e:
            print(e)
            shutil.rmtree(server_directory, ignore_errors=True)
            exit(1)

    version_file = os.path.join(server_directory, "__version__")
    if require_update(version_file):
        try:
            update_server_version()
        except Exception as e:
            print(e)
            if os.path.isfile(version_file):
                os.remove(version_file)
            exit(1)


class DevelopCommand(develop):
    """Post-installation for development mode."""

    def run(self):
        develop.run(self)
        update()


class InstallCommand(install):
    """Post-installation for installation mode."""

    def run(self):
        install.run(self)
        update()


class EggInfoCommand(egg_info):
    """Post-installation for egg info mode."""

    def run(self):
        egg_info.run(self)
        update()


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
        "develop": DevelopCommand,
        "install": InstallCommand,
        "egg_info": EggInfoCommand,
    },
    python_requires=">=3.7",
)
