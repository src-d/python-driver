import os

"""
Very small code to test that the python_driver works. Tests for
version detection and AST generation should be done on
the pydetector module tests so this is keept simple
so we can tests the receive/reply cycle of the main module
"""

def main():
    files = os.listdir('.')
    if files:
        print('Yep, there are files')

    else:
        raise Exception('no files!')

if __name__ == '__main__':
    main()
