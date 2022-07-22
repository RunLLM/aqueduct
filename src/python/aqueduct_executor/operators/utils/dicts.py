class ObjectResult:
    name: str
    succeeded: bool

    def __init__(self, name: str, succeeded: bool):
        self.name = name
        self.succeeded = succeeded
