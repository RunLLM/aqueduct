# Unlike nltk, the backend does not have pyplot pre-installed, so we import this to test that
# installation-on-the-fly works.
import matplotlib.pyplot as plt

from aqueduct import op


@op
def noop_model_with_requirements_file(df):
    plt.show()
    return df
