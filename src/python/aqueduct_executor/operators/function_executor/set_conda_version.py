import os
import sys

PYTHON_VERSION_FILE = "python_version.txt"

# Conda environments
PY_37 = "py37"
PY_38 = "py38"
PY_39 = "py39"
PY_310 = "py310"


if __name__ == "__main__":
    # Change to the path where PYTHON_VERSION_FILE resides.
    # The path is input from the calling bash script.
    os.chdir(sys.argv[1])

    if os.path.isfile(PYTHON_VERSION_FILE):
        with open(PYTHON_VERSION_FILE, "r") as f:
            full_python_version = f.read()
            python_version = full_python_version.split(" ")[-1].strip()
            if python_version == "3.10":
                conda_env = PY_310
            elif python_version == "3.9":
                conda_env = PY_39
            elif python_version == "3.8":
                conda_env = PY_38
            elif python_version == "3.7":
                conda_env = PY_37
            else:
                raise Exception("Unsupported Python version: %s" % python_version)
    else:
        # For backwards compatabiliy. Ensures that workflows without a python_version.txt still work.
        conda_env = PY_38

    # The output of the print statement to stdout is captured by the calling bash script into a variable,
    # so we should not include any other print statements in this Python script.
    print(conda_env)
