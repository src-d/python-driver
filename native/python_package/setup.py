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
    license="GPL3",
    test_suite="test",
    url="https://github.com/bblfsh/python-driver",
    download_url="https://github.com/bblfsh/python-driver/archive/master.zip",
    author="Juanjo Alvarez",
    author_email="juanjo@sourced.tech",
    packages=find_packages(exclude=["tests"]),
    entry_points={
        "console_scripts": [
            "python_driver = python_driver.cli:main"
        ]
    },
    install_requires=[
        "pydetector-bblfsh==0.11.2"
    ],
    classifiers=[
        "Development Status :: 4 - Beta",
        "Environment :: Console",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Compilers",
        "Topic :: Software Development :: Disassemblers",
        "License :: OSI Approved :: GPL3 License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3.6"
    ]
)
