from typing import Any, Callable, Union

import numpy as np

UserFunction = Callable[..., Any]
Number = Union[int, float, np.number]
MetricFunction = Callable[..., Number]
CheckFunction = Callable[..., Union[bool, np.bool_]]
