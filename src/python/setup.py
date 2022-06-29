from setuptools import find_packages, setup
from setuptools.command.develop import develop
from setuptools.command.install import install
from pathlib import Path
import os



def update_executable_permissions():
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "server")])
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "executor")])
    execute_command(["chmod", "755", os.path.join(server_directory, "bin", "migrator")])

def download_server_binaries():
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


class PostDevelopCommand(develop):
    """Post-installation for development mode."""
    def run(self):
        develop.run(self)
        print("UI>>>")
        update_ui_version()
        update_server_version()

class PostInstallCommand(install):
    """Post-installation for installation mode."""
    def run(self):
        install.run(self)
        update_ui_version()
        update_server_version()

install_requires = open("requirements.txt").read().strip().split("\n")

readme_path = Path(os.environ["PWD"], "../../README.md")
long_description = open(readme_path).read()

setup(
    name="aqueduct-ml",
    version="0.0.3",
    install_requires=install_requires,
    scripts=["bin/aqueduct.py"],
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
        'develop': PostDevelopCommand,
        'install': PostInstallCommand,
    },
    python_requires=">=3.7",
)
