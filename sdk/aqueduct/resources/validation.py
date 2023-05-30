from functools import wraps
from typing import Any, Callable

from aqueduct.utils.resource_validation import validate_resource_is_connected

AnyFunc = Callable[..., Any]


def validate_is_connected() -> Callable[[AnyFunc], AnyFunc]:
    """This decorator, which must be used on a Resource class method,
    ensures that the resource is connected before allowing the method to be called."""

    def decorator(method: AnyFunc) -> Callable[[AnyFunc], AnyFunc]:
        @wraps(method)
        def wrapper(self: Any, *args: Any, **kwargs: Any) -> Any:
            validate_resource_is_connected(self.name(), self._metadata.exec_state)
            return method(self, *args, **kwargs)

        return wrapper

    return decorator
