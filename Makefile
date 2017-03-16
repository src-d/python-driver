test-native-internal:
	cd native/python_package/test; \
	python3 -m unittest discover

build-native-internal:
	cd native/python_package/; \
	pip3 install -U --user .

include .sdk/Makefile
