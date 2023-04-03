import inspect
import os
import shutil
import sys
import tempfile
import uuid
from pathlib import Path
from typing import List, Optional, Union

import cloudpickle as pickle
import pkg_resources
from aqueduct.error import (
    InternalAqueductError,
    InvalidDependencyFilePath,
    InvalidFunctionException,
    RequirementsMissingError,
    ReservedFileNameException,
)
from aqueduct.logger import logger
from aqueduct.type_annotations import CheckFunction, MetricFunction, UserFunction

MODEL_FILE_NAME = "model.py"
MODEL_PICKLE_FILE_NAME = "model.pkl"
PYTHON_VERSION_FILE_NAME = "python_version.txt"
CONDA_VERSION_FILE_NAME = "conda_version.txt"
RESERVED_FILE_NAMES = [
    MODEL_FILE_NAME,
    MODEL_PICKLE_FILE_NAME,
    PYTHON_VERSION_FILE_NAME,
    CONDA_VERSION_FILE_NAME,
]
REQUIREMENTS_FILE = "requirements.txt"
BLACKLISTED_REQUIREMENTS = [
    "aqueduct_ml",
    "aqueduct_sdk",
    "aqueduct-ml",
    "aqueduct-sdk",
]

DEFAULT_OP_CLASS_NAME = "Function"
DEFAULT_OP_METHOD_NAME = "predict"
_FILE_TEMPLATE = """
import cloudpickle as cp

class {class_name}:
    def __init__(self):
        with open("./model.pkl", "rb") as f:
            self.func = cp.load(f)

    def {method_name}(self, *args):
        return self.func(*args)
"""


def _op_file_content(
    class_name: str = DEFAULT_OP_CLASS_NAME, method_name: str = DEFAULT_OP_METHOD_NAME
) -> str:
    return _FILE_TEMPLATE.format(class_name=class_name, method_name=method_name)


def serialize_function(
    func: Union[UserFunction, MetricFunction, CheckFunction],
    op_name: str,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> bytes:
    """
    Takes a user-defined function and packages it into a zip file structure expected by the backend.

    Arguments:
        func:
            The function to package
        op_name:
            The name of the function operator to package.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
        requirements:
            Defines the python package requirements that this function will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise  RequirementsFileMissingError exception will be raised.

    Returns:
        filepath of zip file in string format
    """
    dir_path = None
    try:
        dir_path = _make_temp_dir()
        _package_files_and_requirements(
            func, os.path.join(os.getcwd(), dir_path), file_dependencies, requirements
        )

        # Figure out the python version
        python_version = ".".join((str(x) for x in sys.version_info[:2]))
        with open(os.path.join(dir_path, PYTHON_VERSION_FILE_NAME), "w") as f:
            f.write(python_version)

        with open(os.path.join(dir_path, MODEL_FILE_NAME), "w") as model_file:
            model_file.write(_op_file_content())
        with open(os.path.join(dir_path, MODEL_PICKLE_FILE_NAME), "wb") as f:
            pickle.dump(func, f)

        # Write function source code to file
        source_file = "{}.py".format(op_name)
        with open(os.path.join(dir_path, source_file), "w") as f:
            try:
                source = inspect.getsource(func)
            except Exception:
                source = "unknown"
            f.write(source)

        zip_file_path = get_zip_file_path(dir_path)
        _make_archive(dir_path, zip_file_path)
        return open(zip_file_path, "rb").read()
    finally:
        if dir_path:
            delete_zip_folder_and_file(dir_path)


def _package_files_and_requirements(
    func: Union[UserFunction, MetricFunction, CheckFunction],
    dir_path: str,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> None:
    """
    Populates the given dir_path directory with all the file dependencies and requirements.txt.

    Arguments:
        func:
            User-defined function to package
        dir_path:
            Absolute path of directory we'll be using
        file_dependencies:
            A list of relative paths to files that the function needs to access.
        requirements:
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those.
    """
    if not file_dependencies:
        file_dependencies = []

    current_directory_path = os.getcwd()

    func_filepath = inspect.getsourcefile(func)
    if not func_filepath:
        raise Exception("Unable to find source file of function.")
    # In Python3.8, `inspect.getsourcefile` only returns the file's relative path,
    # so we need the line below to get the absolute path.
    func_filepath = os.path.abspath(func_filepath)
    func_dirpath = os.path.dirname(func_filepath)

    # We check if the directory `func_dirpath` exists. If not, this means `func` is from within a
    # Jupyter notebook that the user is currently running, so we don't switch the working directory.
    # The goal of switching the working directory is that if a user specifies relative paths
    # in `file_dependencies` and if `func` is imported from a Python script located in another
    # directory, we can locate them.
    if os.path.isdir(func_dirpath):
        os.chdir(func_dirpath)

    try:
        for file_index, file_path in enumerate(file_dependencies):
            if file_path in RESERVED_FILE_NAMES:
                # If the user uploads a `model.py` file as a dependency, we will error out.
                raise ReservedFileNameException(
                    "%s is a reserved file name in our system. Please rename your file. "
                    % file_path
                )
            if not os.path.exists(file_path):
                raise InvalidFunctionException("File %s does not exist" % file_path)

            if not os.path.abspath(file_path).startswith(os.getcwd()):
                raise InvalidDependencyFilePath(
                    "File %s cannot be outside of the directory containing the function" % file_path
                )

            dstfolder = os.path.dirname(os.path.join(dir_path, file_path))
            if not os.path.exists(dstfolder):
                os.makedirs(dstfolder)
            shutil.copy(file_path, os.path.join(dir_path, file_path))

        # This is the absolute path to the requirements file we are sending to the backend.
        packaged_requirements_path = os.path.join(dir_path, REQUIREMENTS_FILE)
        if requirements is not None:
            # The operator has a custom requirements specification.
            assert isinstance(requirements, str) or all(
                isinstance(req, str) for req in requirements
            )

            if isinstance(requirements, str):
                if os.path.exists(requirements):
                    logger().info(
                        "Installing requirements found at {path}".format(path=requirements)
                    )
                    shutil.copy(requirements, packaged_requirements_path)
                else:
                    raise FileNotFoundError(
                        "Requirements file provided at %s does not exist." % requirements
                    )
            else:
                # User has given us a list of pip requirement strings.
                with open(packaged_requirements_path, "x") as f:
                    f.write("\n".join(requirements))

        elif os.path.exists(REQUIREMENTS_FILE):
            # There exists a workflow-level requirements file (need to reside in the same directory as the function).
            shutil.copy(REQUIREMENTS_FILE, packaged_requirements_path)

        else:
            # Do not infer package dependencies from environment.
            raise RequirementsMissingError(
                "A valid requirements.txt file must be provided "
                "and must be in the same directory as the function "
                "definition or alternatively add the requirements "
                "directly in the @op/@metric/@check decorator"
            )

        # Prune out any blacklisted requirements.
        _filter_out_blacklisted_requirements(packaged_requirements_path)

        _add_cloudpickle_to_requirements(packaged_requirements_path)

    finally:
        os.chdir(current_directory_path)


def _filter_out_blacklisted_requirements(packaged_requirements_path: str) -> None:
    """Opens the requirements.txt file and removes any packages that we don't support."""
    with open(packaged_requirements_path, "r") as f:
        req_lines = f.readlines()

    with open(packaged_requirements_path, "w") as f:
        for line in req_lines:
            if any(blacklisted_req in line for blacklisted_req in BLACKLISTED_REQUIREMENTS):
                continue
            f.write(line)


def _add_cloudpickle_to_requirements(packaged_requirements_path: str) -> None:
    """
    Regardless of how we detect dependencies, we must include cloudpickle (with client's version
    number) as a requirement because the server needs to install the same version of cloudpickle as
    the client.

    If the user-specified requirements file already contains a cloudpickle entry and the version
    number matches, the installation process will still succeed. If there is a mismatch, the installation
    will fail and the user should fix the version number to match the version installed on the client.
    """
    with open(packaged_requirements_path, "a") as f:
        cloudpickle_requirement = (
            "\ncloudpickle==%s" % pkg_resources.get_distribution("cloudpickle").version
        )
        f.write(cloudpickle_requirement)


def get_zip_file_path(dir_name: str) -> str:
    return dir_name + ".zip"


def delete_zip_folder_and_file(dir_name: str) -> None:
    zip_file_path = get_zip_file_path(dir_name)

    if os.path.isfile(zip_file_path):
        os.remove(zip_file_path)

    if os.path.exists(dir_name):
        shutil.rmtree(dir_name)


def _make_archive(source: str, destination: str) -> None:
    """Creates zip file from source directory to destination file."""
    base = os.path.basename(destination)
    split_base = base.split(".")
    name = Path(".".join(split_base[:-1])).name
    format = split_base[-1]
    archive_from = os.path.dirname(source)
    if not archive_from:
        archive_from = "."
    archive_to = os.path.basename(source.strip(os.sep))
    shutil.make_archive(name, format, archive_from, archive_to)
    shutil.move("%s.%s" % (name, format), destination)


def _make_temp_dir() -> str:
    """
    Create a unique, temporary directory in the local filesystem and returns the path.
    """
    dir_path = None
    created = False
    # Try to create the directory. If it already exists, try again with a new name.
    while not created:
        dir_path = Path(tempfile.gettempdir()) / str(uuid.uuid4())
        try:
            os.mkdir(dir_path)
            created = True
        except FileExistsError:
            pass

    assert dir_path is not None
    return str(dir_path)
