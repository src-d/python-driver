@decorator(param1, param2)
def somefunction(arg1, arg2, arg3):
    """
    This is the doc
    """
    arg1 = arg2 + arg3
    if arg1 > 1:
        yield arg1
    return arg2

l = [4, 8]
somefunction(1, *l)
