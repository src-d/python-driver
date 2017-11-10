DEV_DEPS ?= native/dev_deps

test-native-internal:
	pip3 install --user ${DEV_DEPS}/python-pydetector/ || pip3 install --user pydetector-bblfsh
	cd native/python_package/test && \
	python3 -m unittest discover

build-native-internal:
	pip3 install --user ${DEV_DEPS}/python-pydetector/ || pip3 install --user pydetector-bblfsh
	cd native/python_package/ && \
	pip3 install -U --user .
	cp native/sh/native.sh $(BUILD_PATH)/bin/native;
	chmod +x $(BUILD_PATH)/bin/native


include .sdk/Makefile
