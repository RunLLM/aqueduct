import sys
import os

from setuptools import find_packages, setup


install_requires = open("requirements.txt").read().strip().split("\n")

setup(
    name="aqueduct-ml",
    version="0.0.1",
    install_requires=install_requires,
    scripts=['bin/aqueduct'],
    packages=find_packages(),
    description="Aqueduct OSS backend.",
    url="https://www.aqueducthq.com/",
    license="Apache License 2.0",
    author="Aqueduct, Inc.",
    author_email="hello@aqueducthq.com",
    classifiers=[
        "Programming Language :: Python :: 3",
    ],
    python_requires=">=3.7,<=3.10",
)
