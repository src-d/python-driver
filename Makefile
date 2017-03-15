test-native-internal:
	cd native/python_package/test; \
	python3 -m unittest discover

build-native-internal:
	cd native/python_packagexx/; \
	pip3 install -U .

include .sdk/Makefile
