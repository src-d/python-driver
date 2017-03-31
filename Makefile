test-native-internal:
	cd native/python_package/test; \
	python3 -m unittest discover

build-native-internal:
	cd native/python_package/; \
	pip3 install -U --user .
	cp native/sh/native.sh $(BUILD_PATH)/native;
	chmod +x $(BUILD_PATH)/native


include .sdk/Makefile
