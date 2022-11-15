from typing import Any

from aqueduct.decorator import _convert_memory_string_to_mbs
from aqueduct.error import InvalidUserArgumentException


def test_convert_memory_string_to_mbs():
    def _run_test(input: str, expected: int, err_msg: str = ""):
        """`err_msg` only needs to be a substring of the exception message."""
        try:
            output = _convert_memory_string_to_mbs(input)
            assert err_msg == "", "Exception was expected."
            assert output == expected
        except InvalidUserArgumentException as e:
            assert err_msg != "", "Exception was not expected."
            assert err_msg in str(e)
        except Exception:
            assert False, "Wrong exception type was raised."

    _run_test("150MB", 150)
    _run_test("150Mb", 150)
    _run_test("150mb", 150)
    _run_test("150GB", 150 * 1000)
    _run_test("150Gb", 150 * 1000)
    _run_test("150gb", 150 * 1000)
    _run_test("150GB       ", 150 * 1000)
    _run_test("150   GB       ", 150 * 1000)
    _run_test("   150 GB  ", 150 * 1000)

    _run_test("1", -1, "not long enough")
    _run_test("150", -1, "must have a suffix that is one of")
    _run_test("150de", -1, "must have a suffix that is one of")
    _run_test("-150 MB", -1, "must be a positive integer")
    _run_test("abcMB", -1, "must be a positive integer")
    _run_test("abc150MB", -1, "must be a positive integer")
