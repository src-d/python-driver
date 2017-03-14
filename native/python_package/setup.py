# -*- coding: utf-8 -*-

from setuptools import setup, find_packages
from codecs import open
from os import path

here = path.abspath(path.dirname(__file__))

description = "Python driver for the bblfsh project"

main_ns = {}
with open("python_driver/version.py") as ver_file:
    exec(ver_file.read(), main_ns)

version = main_ns['__version__']
setup(
    name="python_driver",
    version=version,
    description=description,
    long_description=description,
    license="MIT",
    test_suite="test",
    # TODO: change this to the bblfsh directory
    url="https://github.com/juanjux/python-driver",
    # download_url = "https://github.com/juanjux/python_driver/archive/%s.tar.gz" % version,
    download_url="https://github.com/juanjux/python_driver/archive/astcomments_profiling.tar.gz",
    author="Juanjo Alvarez",
    author_email="juanjo@sourced.tech",
    packages=find_packages(exclude=["tests"]),
    entry_points={
        "console_scripts": [
            "python_driver = python_driver.cli:main"
        ]
    },
    install_requires=[
        "msgpack-python==0.4.8",
        "pydetector==0.3.14"
    ],
    classifiers=[
        "Development Status :: 4 - Beta",
        "Environment :: Console",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Compilers",
        "Topic :: Software Development :: Disassemblers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3.6"
    ]
)
