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


def op_file_content(
    class_name: str = DEFAULT_OP_CLASS_NAME, method_name: str = DEFAULT_OP_METHOD_NAME
) -> str:
    return _FILE_TEMPLATE.format(class_name=class_name, method_name=method_name)
