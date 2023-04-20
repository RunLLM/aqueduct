from typing import Any, Callable

from aqueduct.utils.integration_validation import validate_integration_is_connected

AnyFunc = Callable[..., Any]


def validate_is_connected() -> Callable[[AnyFunc], AnyFunc]:
    """This decorator, which must be used on an Integration class method,
    ensures that the integration is connected before allowing the method to be called."""

    def decorator(method: AnyFunc) -> Callable[[AnyFunc], AnyFunc]:
        def wrapper(self: Any, *args: Any, **kwargs: Any) -> Any:
            validate_integration_is_connected(self.name(), self._metadata.exec_state)
            return method(self, *args, **kwargs)

        return wrapper

    return decorator
