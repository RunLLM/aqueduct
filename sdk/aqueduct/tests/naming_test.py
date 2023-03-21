from aqueduct.utils.naming import bump_artifact_suffix


def test_bump_artifact_suffix():
    input_to_expected_output = {
        "foo": "foo (1)",
        "foo (1)": "foo (2)",
        "foo)": "foo) (1)",
        "foo (1) (2)": "foo (1) (3)",
        "foo (2) (1": "foo (2) (1 (1)",
        "foo(1)": "foo(1) (1)",
        "foo (123)": "foo (124)",
    }

    for input, expected_output in input_to_expected_output.items():
        assert bump_artifact_suffix(input) == expected_output
