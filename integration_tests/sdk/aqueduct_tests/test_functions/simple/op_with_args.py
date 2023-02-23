from aqueduct import op


# Method is defined here because, if defined inline within a test file, cloudpickle
# will pick up `import pytest`, which does not work when testing against our Conda integration,
# for example.
@op
def foo_with_args(*args):
    return list(args)