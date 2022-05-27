from gettext import install
import os
import setuptools
import sys


req_file_name = "requirements/python-%s-%s.txt" % (sys.version_info[0], sys.version_info[1])

if os.path.exists(req_file_name):
    install_requires = open(req_file_name).read().strip().split("\n")
else:
    raise Exception(
        "Python Version %s.%s not supported" % (sys.version_info[0], sys.version_info[1])
    )

setuptools.setup(
    name="aqueduct-sdk",
    version="0.0.1",
    author="Aqueduct, Inc.",
    author_email="hello@aqueducthq.com",
    description="Python SDK for the Aqueduct prediction infrastructure.",
    url="https://github.com/aqueducthq/aqueduct",
    license="Apache License 2.0",
    packages=setuptools.find_packages(),
    install_requires=install_requires,
    setup_requires=["numpy", "cython", "packaging"],
    classifiers=[
        "Programming Language :: Python :: 3",
    ],
    python_requires=">=3.7,<=3.10",
)
