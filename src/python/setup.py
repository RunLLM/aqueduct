import os
from pathlib import Path

from setuptools import find_packages, setup

install_requires = open("requirements.txt").read().strip().split("\n")

readme_path = Path(os.environ["PWD"], "../../README.md")
long_description = open(readme_path).read()

setup(
    name="aqueduct-ml",
    version="0.1.3",
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
    python_requires=">=3.7",
)
