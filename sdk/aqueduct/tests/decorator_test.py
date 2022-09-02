import io
import os
import zipfile

import pytest
from aqueduct.error import InvalidDependencyFilePath, ReservedFileNameException
from aqueduct.utils import (
    _package_files_and_requirements,
    delete_zip_folder_and_file,
    serialize_function,
)

from .test_files.python_function.python_function import python_function
from .test_files.python_function.test_dependency_folder.helper_function import helper_function


def test_packaging_files():
    try:
        os.mkdir("test_function")
        _package_files_and_requirements(
            python_function,
            dir_path=os.path.join(os.getcwd(), "test_function/"),
            file_dependencies=["./test_dependency_folder/helper_function.py"],
        )

        assert not os.path.exists("./test_function/test_function.py")
        assert os.path.exists("./test_function/test_dependency_folder/helper_function.py")

    finally:
        delete_zip_folder_and_file("test_function")


def test_invalid_dependency_file():
    with pytest.raises(InvalidDependencyFilePath):
        _package_files_and_requirements(
            helper_function,
            dir_path=os.path.join(os.getcwd(), "test_invalid_function/"),
            file_dependencies=["../python_function.py"],
        )


def test_handle_reserved_file_dependencies():
    with pytest.raises(ReservedFileNameException):
        _package_files_and_requirements(
            python_function,
            dir_path=os.path.join(os.getcwd(), "test_function"),
            file_dependencies=["model.py"],
        )

    with pytest.raises(ReservedFileNameException):
        _package_files_and_requirements(
            python_function,
            dir_path=os.path.join(os.getcwd(), "test_function"),
            file_dependencies=["conda_version.txt", "file.py"],
        )


def test_serialize_function():
    initial_state = set(os.listdir())
    dependencies = ["./test_dependency_folder/helper_function.py"]
    op_name = "helper_fn"
    zip_file = serialize_function(
        func=python_function,
        op_name=op_name,
        file_dependencies=dependencies,
    )
    final_state = set(os.listdir())
    assert initial_state == final_state
    assert zip_file is not None
    zip_buffer = io.BytesIO(zip_file)
    zip_file = zipfile.ZipFile(zip_buffer, "r", zipfile.ZIP_DEFLATED, False)
    files = [
        "helper_function.py",
        "python_version.txt",
        "model.py",
        "model.pkl",
        "requirements.txt",
        "{}.py".format(op_name),
    ]
    zip_files = [name.split("/")[-1] for name in zip_file.namelist() if name.split("/")[-1]]
    assert set(zip_files) == set(files)
