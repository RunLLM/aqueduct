import os
from pathlib import Path

from setuptools import find_packages, setup

version = open("version").read()
if not version:
    raise Exception("Version file must contain a valid version string.")

install_requires = open("requirements.txt").read().strip().split("\n")
# We expect the SDK version to be always consistent with the executor version
install_requires.append(f"aqueduct-sdk=={version}")

readme_path = Path(os.environ["PWD"], "../../README.md")
long_description = open(readme_path).read()

setup(
    name="aqueduct-ml",
    version=version,
    install_requires=install_requires,
    scripts=["bin/aqueduct"],
    packages=find_packages(),
    description="The control center for ML in the cloud",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://www.aqueducthq.com/",
    license="Apache License 2.0",
    author="Aqueduct, Inc.",
    author_email="hello@aqueducthq.com",
    classifiers=[
        "Programming Language :: Python :: 3",
    ],
    python_requires=">=3.7",
)
