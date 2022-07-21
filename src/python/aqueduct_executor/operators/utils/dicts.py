class TableOutput(dict):
    def __init__(self, name:str, succeeded: bool):
        dict.__init__(self, name=name, succeeded=succeeded)