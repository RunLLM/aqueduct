import sys

import pandas as pd
import pytest
from aqueduct.error import AqueductError, InvalidUserArgumentException
from constants import SENTIMENT_SQL_QUERY
from transformers_model.model import sentiment_prediction_using_transformers
from utils import get_integration_name

from aqueduct import infer_requirements, op

INVALID_REQUIREMENTS_PATH = "~/random.txt"
VALID_REQUIREMENTS_PATH = "transformers_model/requirements.txt"


def _transformers_package_exists():
    try:
        pass
    except ImportError:
        return False
    return True


def _run_shell_command(cmd: str):
    import subprocess

    process = subprocess.Popen(
        cmd,
        shell=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    stdout_raw, stderr_raw = process.communicate()
    if len(stdout_raw) > 0:
        print(stdout_raw)
    if len(stderr_raw) > 0:
        print(stderr_raw)


def _uninstall_transformers_package():
    print("Uninstalling `transformers` package.")
    _run_shell_command(f"{sys.executable} -m pip uninstall -y transformers")


def _install_transformers_package():
    print("Installing `transformers` package.")
    _run_shell_command(f"{sys.executable} -m pip install transformers")


def _check_infer_requirements(transformers_exists: bool):
    """Runs our supplied `infer_requirements` method and checks for the presence of `transformers`."""
    inferred_reqs = infer_requirements()
    if transformers_exists:
        assert any(req_str.startswith("transformers") for req_str in inferred_reqs)
    else:
        assert all(not req_str.startswith("transformers") for req_str in inferred_reqs)


@op()
def sentiment_prediction_without_reqs_path(reviews: pd.DataFrame) -> pd.DataFrame:
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


@pytest.mark.requirements
def test_infer_requirements(client):
    if _transformers_package_exists():
        _uninstall_transformers_package()

    # If no requirements are supplied, our inference will not pick up the transformers package.
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    with pytest.raises(AqueductError):
        sentiment_prediction_without_reqs_path(table)

    _check_infer_requirements(transformers_exists=False)
    _install_transformers_package()

    # After installing the transformers package in the environment, check that we now
    # are inferring the `transformers` package. There's not a great way of checking this exactly,
    # since the SDK and backend are expected to share the same python environment currently.
    # TODO(ENG-1471): Once installation isolation is provided, we can rewrite this to test case
    #  to simply check that preview works.
    _check_infer_requirements(transformers_exists=True)


@op(requirements=INVALID_REQUIREMENTS_PATH)
def sentiment_prediction_with_invalid_reqs_path(table: pd.DataFrame) -> pd.DataFrame:
    return table


@op(requirements=VALID_REQUIREMENTS_PATH)
def sentiment_prediction_with_valid_reqs_path(reviews: pd.DataFrame) -> pd.DataFrame:
    """This uses the requirements.txt in the transformers_model/ subdirectory to install transformers."""
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


@pytest.mark.requirements
def test_requirements_installation_from_path(client):
    if _transformers_package_exists():
        _uninstall_transformers_package()

    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)

    # Check that no an invalid path fails.
    with pytest.raises(FileNotFoundError):
        sentiment_prediction_with_invalid_reqs_path(table)

    valid_path_table = sentiment_prediction_with_valid_reqs_path(table)
    assert valid_path_table.get().shape[0] == 100


@op(requirements=["transformers==4.21.0"])
def sentiment_prediction_with_string_requirements(reviews):
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


@pytest.mark.requirements
def test_requirements_installation_from_strings(client):
    # TODO(test the list if not well-formed.)
    if _transformers_package_exists():
        _uninstall_transformers_package()

    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    valid_path_table = sentiment_prediction_with_string_requirements(table)
    assert valid_path_table.get().shape[0] == 100


@pytest.mark.requirements
def test_default_requirements_installation(client):
    """
    This uses the decorated function in the transformers_model/ subdirectory, which already has a requirements.txt
    that it should be installing.
    """
    if _transformers_package_exists():
        _uninstall_transformers_package()

    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    valid_path_table = sentiment_prediction_using_transformers(table)
    assert valid_path_table.get().shape[0] == 100


@pytest.mark.requirements
def test_requirements_invalid_arguments(client):

    with pytest.raises(InvalidUserArgumentException):

        @op(requirements=123)
        def fn_wrong_requirements_type(table):
            return table
