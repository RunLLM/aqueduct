from setuptools import setup, find_packages

install_requires = open("requirements.txt").read().strip().split("\n")

setup(
    name="aqueduct-llm",
    version="0.0.0",
    install_requires=install_requires,
    packages=find_packages(),
    description="Aqueduct LLM Package",
    long_description="This package allows you to integrate LLMs into your Aqueduct machine learning pipelines with a single line of code.",
    long_description_content_type="text/markdown",
    url="https://www.aqueducthq.com/",
    license="Apache License 2.0",
    author="Aqueduct, Inc.",
    author_email="hello@aqueducthq.com",
    classifiers=[
        "Programming Language :: Python :: 3",
    ],
    python_requires=">=3.8",
)
